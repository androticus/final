package core

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"unicode/utf8"
)

const (
	errorStringMaxRunes   = 40
	faultStackSnapshotMax = 16
	faultStackReportHead  = 8
	faultStackReportTail  = 2
)

// RunFault is a structured snapshot after a VM panic (computation state is undefined).
type RunFault struct {
	Kind        string
	Message     string
	TaskName    string
	ReturnStack []string // lines "index: …" (0 = RS bottom); "*:" = fault site
	DataStack   []string // lines "index: …" (0 = stack bottom)
	TempStack   []string // lines "index: …" (0 = stack bottom)
	XStack      []string // lines "index: …" (0 = stack bottom)
	YStack      []string // lines "index: …" (0 = stack bottom)
	ZStack      []string // lines "index: …" (0 = stack bottom)
	OStack      []string // lines "index: …" (0 = stack bottom)
	LastOp      OpCode
	IP          IP
	Defining    bool
	EmittingCT  bool
}

func (f *RunFault) Error() string {
	if f == nil {
		return "<nil RunFault>"
	}
	return f.Message
}

const (
	rsLegendFaultSite = "* = fault site"
	rsLegendCurrent   = "* = current"
)

const stackReportBodyTemplateText = `Return Stack (tos first, {{.RSLegend}}):
{{range .ReturnStack}}  {{.}}
{{end}}Data Stack (tos first):
{{range .DataStack}}  {{.}}
{{end}}Temp Stack (tos first):
{{range .TempStack}}  {{.}}
{{end}}X Stack (tos first):
{{range .XStack}}  {{.}}
{{end}}Y Stack (tos first):
{{range .YStack}}  {{.}}
{{end}}Z Stack (tos first):
{{range .ZStack}}  {{.}}
{{end}}O Stack (tos first):
{{range .OStack}}  {{.}}
{{end}}`

var stackReportBodyTemplate = template.Must(template.New("stackReport").Parse(stackReportBodyTemplateText))

type faultReportView struct {
	RSLegend    string
	Message     string
	ReturnStack []string
	DataStack   []string
	TempStack   []string
	XStack      []string
	YStack      []string
	ZStack      []string
	OStack      []string
}

// FormatErrorCell formats a stack cell for fault reports (never panics; elides long strings).
func FormatErrorCell(v any) string {
	if v == nil {
		return "nil"
	}
	if ns, ok := v.(NaturalStringer); ok {
		return ns.NaturalString()
	}
	switch x := v.(type) {
	case String:
		return elideErrorString(string(x))
	case Bool:
		if x {
			return "true"
		}
		return "false"
	case Int:
		return strconv.Itoa(int(x))
	case Uint:
		// Affix matches natural string / literal syntax so numeric type is visible in faults.
		return strconv.FormatUint(uint64(x), 10) + "u"
	case Float:
		return formatNaturalFloat(x)
	case *AnyList:
		if x == nil {
			return "List(nil)"
		}
		return fmt.Sprintf("List(len=%d)", x.Len())
	case *TextReader:
		return "TextReader"
	case *TextWriter:
		return "TextWriter"
	default:
		return fmt.Sprintf("<%T>", v)
	}
}

func elideErrorString(s string) string {
	if s == "" {
		return `""`
	}
	if utf8.RuneCountInString(s) <= errorStringMaxRunes {
		return strconv.Quote(s)
	}
	var b strings.Builder
	b.WriteByte('"')
	n := 0
	for _, r := range s {
		if n >= errorStringMaxRunes {
			b.WriteString("…")
			break
		}
		b.WriteRune(r)
		n++
	}
	b.WriteByte('"')
	return b.String()
}

func elideStackLines(lines []string, headN, tailM int) []string {
	if len(lines) <= headN+tailM {
		return lines
	}
	out := make([]string, 0, headN+1+tailM)
	out = append(out, lines[:headN]...)
	out = append(out, "...")
	out = append(out, lines[len(lines)-tailM:]...)
	return out
}

func faultReportStackLines(lines []string) []string {
	if len(lines) == 0 {
		return nil
	}
	return elideStackLines(lines, faultStackReportHead, faultStackReportTail)
}

func faultReportCellStackLines(lines []string) []string {
	if len(lines) == 0 {
		return []string{"(empty)"}
	}
	return faultReportStackLines(lines)
}

// RunFaultReportHook is invoked after a fault report is written to stdout (e.g. GUI
// appendResponse). Nil outside hosts that register one. Stderr is reserved for user $$.
var RunFaultReportHook func(fault *RunFault)

// WriteRunFaultReport renders a RunFault to stdout via text/template, then notifies Hook.
func WriteRunFaultReport(rp *RunProvider, fault *RunFault) {
	if fault == nil {
		return
	}
	if rp != nil && rp.Stdout != nil {
		rp.Stdout.WriteString(FormatRunFaultReport(fault))
		flushTextWrite(rp.Stdout)
	}
	if RunFaultReportHook != nil {
		RunFaultReportHook(fault)
	}
}

// collectStackSnapshotLines gathers RS/DS/TS/X/Y/Z/O lines for reports (*$ axiom and RunFault).
func collectStackSnapshotLines(rp *RunProvider, execDC *DictionaryContext, ip IP, lastOp OpCode) (returnStack, dataStack, tempStack, xStack, yStack, zStack, oStack []string) {
	if rp == nil {
		return nil, nil, nil, nil, nil, nil, nil
	}
	return buildFaultReturnStack(rp.ReturnStack, execDC, ip, lastOp),
		buildFaultStackSnapshot(rp.DataStack),
		buildFaultStackSnapshot(rp.TempStack),
		buildFaultStackSnapshot(rp.XStack),
		buildFaultStackSnapshot(rp.YStack),
		buildFaultStackSnapshot(rp.ZStack),
		buildFaultStackSnapshot(rp.OStack)
}

func stackReportViewFromLines(rsLegend string, rs, ds, ts, xs, ys, zs, osLines []string) faultReportView {
	return faultReportView{
		RSLegend:    rsLegend,
		ReturnStack: faultReportStackLines(rs),
		DataStack:   faultReportCellStackLines(ds),
		TempStack:   faultReportCellStackLines(ts),
		XStack:      faultReportCellStackLines(xs),
		YStack:      faultReportCellStackLines(ys),
		ZStack:      faultReportCellStackLines(zs),
		OStack:      faultReportCellStackLines(osLines),
	}
}

func renderStackReportBody(view faultReportView) string {
	var buf bytes.Buffer
	if err := stackReportBodyTemplate.Execute(&buf, view); err != nil {
		return fmt.Sprintf("(stack report template error: %v)\n", err)
	}
	return buf.String()
}

// FormatStackSnapshotReport renders RS/DS/TS like a fault report (no panic message).
// Used by *$; RS legend is "* = current" (not fault site).
func FormatStackSnapshotReport(rp *RunProvider, execDC *DictionaryContext, ip IP, lastOp OpCode) string {
	rs, ds, ts, xs, ys, zs, osLines := collectStackSnapshotLines(rp, execDC, ip, lastOp)
	return renderStackReportBody(stackReportViewFromLines(rsLegendCurrent, rs, ds, ts, xs, ys, zs, osLines))
}

// FormatRunFaultReport renders a RunFault as a human-readable report string.
func FormatRunFaultReport(fault *RunFault) string {
	if fault == nil {
		return ""
	}
	view := stackReportViewFromLines(rsLegendFaultSite, fault.ReturnStack, fault.DataStack, fault.TempStack, fault.XStack, fault.YStack, fault.ZStack, fault.OStack)
	body := renderStackReportBody(view)
	if fault.TaskName != "" {
		return fmt.Sprintf("Fatal Error: [%s] %s\n%s", fault.TaskName, fault.Message, body)
	}
	return fmt.Sprintf("Fatal Error: %s\n%s", fault.Message, body)
}

func captureRunFault(t *Task, rp *RunProvider, recovered any, lastOp OpCode, ip IP) *RunFault {
	msg := fmt.Sprint(recovered)
	name := ""
	if t != nil {
		name = t.name
	}
	f := &RunFault{
		Kind:     "panic",
		Message:  msg,
		TaskName: name,
		LastOp:   lastOp,
		IP:       ip,
	}
	if t != nil {
		compiling := t.dc != nil && GetCompileState(t.compilerForTarget())
		f.Defining = compiling
		f.EmittingCT = compiling
	}
	if rp != nil {
		f.ReturnStack, f.DataStack, f.TempStack, f.XStack, f.YStack, f.ZStack, f.OStack = collectStackSnapshotLines(rp, rp.LastExecDC, ip, lastOp)
	}
	return f
}

func formatFaultStackLine(index int, value string) string {
	return fmt.Sprintf("%d: %s", index, value)
}

// buildFaultStackSnapshot renders TOS-first (highest index under the heading).
// Index labels still use 0 = stack bottom. When depth exceeds faultStackSnapshotMax,
// the kept cells are the same topmost window; elision may add "..." before deeper cells.
func buildFaultStackSnapshot(s *Stack[any]) []string {
	if s == nil || s.Len() == 0 {
		return nil
	}
	total := s.Len()
	start := 0
	if total > faultStackSnapshotMax {
		start = total - faultStackSnapshotMax
	}
	out := make([]string, 0, total-start)
	for idx := total - 1; idx >= start; idx-- {
		tosIdx := total - 1 - idx
		out = append(out, formatFaultStackLine(idx, FormatErrorCell(*s.AddrAt(tosIdx))))
	}
	return out
}

// buildFaultReturnStack renders TOS-first: fault site "*:" then RS from top index down.
// Fault IP is labeled in execDC. Each frame's lastIP is labeled in lastDC if set,
// otherwise in the DC of the nested body (same-module nest).
func buildFaultReturnStack(rs *ReturnStackType, execDC *DictionaryContext, faultIP IP, lastOp OpCode) []string {
	total := 0
	if rs != nil {
		total = rs.Len()
	}
	out := make([]string, 0, total+1)
	out = append(out, fmt.Sprintf("*: %s", formatReturnStackEntry(execDC, faultIP, lastOp, "at")))
	currentDC := execDC
	for idx := total - 1; idx >= 0; idx-- {
		tosIdx := total - 1 - idx
		mark := rs.At(tosIdx)
		labelDC := mark.lastDC
		if labelDC == nil {
			labelDC = currentDC
		}
		out = append(out, formatFaultStackLine(idx, formatReturnStackEntry(labelDC, IP(mark.lastIP), OpINVALID, "in")))
		currentDC = labelDC
	}
	return out
}

func dcIsCore(dc *DictionaryContext) bool {
	return dc != nil && (dc == globalDictionaryContext || dc.moduleId == CoreModuleSlot)
}

func dcFaultLabel(dc *DictionaryContext) string {
	if dc == nil {
		return "?"
	}
	if m := ModuleTable.moduleByID(dc.moduleId); m != nil {
		if m.prefix != "" {
			return m.prefix
		}
		if m.name != "" {
			return shortModuleLabel(m.name)
		}
	}
	if dc.name != "" {
		return shortModuleLabel(dc.name)
	}
	return fmt.Sprintf("mod%d", dc.moduleId)
}

// shortModuleLabel keeps "core" / "csv"; strips the random id from "local:<n>".
func shortModuleLabel(name string) string {
	if i := strings.IndexByte(name, ':'); i > 0 {
		return name[:i]
	}
	return name
}

func formatReturnStackEntry(dc *DictionaryContext, ctOffset IP, lastOp OpCode, prefix string) string {
	off := int(ctOffset)
	if dc == nil || dc.ct == nil || dc.dict == nil {
		return fmt.Sprintf("%s <ct:%d>", prefix, off)
	}
	if off < 0 || off >= dc.ct.Len() {
		return fmt.Sprintf("%s <%s ct:%d len:%d>", prefix, dcFaultLabel(dc), off, dc.ct.Len())
	}
	label := resolveWordLabelAtCT(dc, off, lastOp)
	return fmt.Sprintf("%s %s (%s ct:%d)", prefix, label, dcFaultLabel(dc), off)
}

// resolveWordLabelAtCT maps a CT offset to a word name for fault call chains.
// It uses body ranges [colonCode, nextEntry), not "largest colonCode <= offset" (which mis-labels
// repl-loop addresses as WarmBoot because that word's colonCode is 2).
func resolveWordLabelAtCT(dc *DictionaryContext, ctOffset int, lastOp OpCode) string {
	if dc == nil || dc.dict == nil || dc.ct == nil {
		return fmt.Sprintf("<ct:%d>", ctOffset)
	}
	if name := colonWordContainingOffset(dc, ctOffset); name != "" {
		return name
	}
	if lastOp != OpINVALID && lastOp != OpTransition && lastOp != OpStop && lastOp != OpNop {
		for _, w := range dc.dict.byOrdinal {
			if w != nil && w.IsAxiom() && w.opCode == lastOp {
				return w.name
			}
		}
		return OpCodeName(lastOp)
	}
	return fmt.Sprintf("<ct:%d>", ctOffset)
}

func colonWordContainingOffset(dc *DictionaryContext, off int) string {
	if dc == nil || dc.dict == nil || dc.ct == nil {
		return ""
	}
	if dcIsCore(dc) {
		if off >= int(ReservedCTBootBodiesStart) && off < int(bootstrapReplBodyEndCTP) {
			return "repl"
		}
		if off >= int(bootstrapReplBodyEndCTP) && off < dc.ct.Len() {
			words := colonWordsSortedByEntry(dc.dict, int(ReservedCTBootBodiesStart))
			if len(words) == 0 || off < int(words[0].colonCode) {
				return bootWordWarmBoot
			}
		}
	}
	minColon := 0
	if dcIsCore(dc) {
		minColon = int(ReservedCTBootBodiesStart)
	}
	words := colonWordsSortedByEntry(dc.dict, minColon)
	for i, w := range words {
		end := dc.ct.Len()
		if i+1 < len(words) {
			end = int(words[i+1].colonCode)
		}
		if off >= int(w.colonCode) && off < end {
			return w.name
		}
	}
	return ""
}

func colonWordsSortedByEntry(dict *Dictionary, minColon int) []*Word {
	var words []*Word
	for _, w := range dict.byOrdinal {
		if w == nil || w.IsAxiom() || w.colonCode < minColon {
			continue
		}
		if w.name == bootWordColdBoot || w.name == bootWordWarmBoot {
			continue
		}
		words = append(words, w)
	}
	sort.Slice(words, func(i, j int) bool {
		return words[i].colonCode < words[j].colonCode
	})
	return words
}

// warmBootReset clears execution stacks and compile-session debris after a fault.
func warmBootReset(t *Task, rp *RunProvider) {
	clearPendingThrow(t)
	if rp != nil {
		rp.DataStack.Clear()
		rp.ReturnStack.Clear()
		rp.TempStack.Clear()
		rp.XStack.Clear()
		rp.YStack.Clear()
		rp.ZStack.Clear()
		rp.OStack.Clear()
	}
	if t != nil {
		cs := t.compilerForTarget()
		_ = cs.tryRollbackTransaction()
		if t.dc != nil {
			SetCompileState(cs, false)
		}
		ClearCoreByeReq()
	}
}

// bootstrapDispatchWarmReset is the OpTransition handler for WarmBoot recovery.
func bootstrapDispatchWarmReset(f *RunFrame) {
	warmBootReset(f.Task, f.RP)
}

// RunREPLSession runs until OpStop, recovering from panics and resuming at the repl loop head.
func RunREPLSession(t *Task, rp *RunProvider) {
	loopHead := EncodeVirt(ModuleIDThis, int(BootstrapReplLoopHeadCTP()))
	for {
		var fault *RunFault
		func() {
			defer func() {
				if r := recover(); r != nil {
					fault = captureRunFault(t, rp, r, rp.LastOp, rp.LastIP)
					warmBootReset(t, rp)
					WriteRunFaultReport(rp, fault)
					rp.ReturnStack.PushNew(int(ReservedCTBootStop))
					t.ip = loopHead
				}
			}()
			runLoop(t, rp)
		}()
		if fault == nil {
			return
		}
	}
}

const eventLoopRootWordDefault = "guiRoot"

// errEventLoopRestart is panicked by AfterNestedInterpretFault so RunEventLoopSession
// re-enters the trampoline after a nested InterpretSource fault. The host must report
// the fault before calling AfterNestedInterpretFault (this path skips WriteRunFaultReport).
type errEventLoopRestart struct{}

// AfterNestedInterpretFault abandons the current EventLoop#run frame after a nested
// InterpretSource (or equivalent) fault. Call only after the host has reported the
// fault to the user. Panics with a sentinel; RunEventLoopSession warm-boots all value
// stacks and re-enters the trampoline word. Does not remint MethodRefs, recreate the
// EventLoop/pipe, or re-queue the Event already pulled.
func AfterNestedInterpretFault() {
	panic(errEventLoopRestart{})
}

// RunEventLoopSession runs rootWord (a colon that calls EventLoop#run from Final)
// until OpStop, recovering from panics and re-entering that word. doneChan is unset
// so panics surface here. Do not OpInterrupt into EventLoop#run from Go.
func RunEventLoopSession(t *Task, rp *RunProvider, rootWord string) {
	if t == nil || rp == nil {
		panic("RunEventLoopSession: nil Task or RunProvider")
	}
	if rootWord == "" {
		rootWord = eventLoopRootWordDefault
	}
	rp.WithTaskFindScope(t)
	w, _, ok := FindNameInTask(t, rp, String(rootWord))
	if !ok || w == nil || w.colonCode < 0 {
		panic("RunEventLoopSession: missing colon word " + rootWord + " (compile host trampoline first)")
	}
	for {
		var fault *RunFault
		var restart bool
		func() {
			defer func() {
				if r := recover(); r != nil {
					if _, ok := r.(errEventLoopRestart); ok {
						warmBootReset(t, rp)
						restart = true
						return
					}
					fault = captureRunFault(t, rp, r, rp.LastOp, rp.LastIP)
					warmBootReset(t, rp)
					WriteRunFaultReport(rp, fault)
				}
			}()
			rp.ReturnStack.PushNew(int(ReservedCTBootStop))
			t.pendingExecWord = w
			t.ip = ReservedCTBootStop
			runLoop(t, rp)
		}()
		if !restart && fault == nil {
			return
		}
	}
}
