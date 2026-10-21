package core

// HttpServeJob is the per-request bag pushed on a route request Pipe.
// The Go ServeHTTP stub waits on reply; Final signals done by pushing a token.
type HttpServeJob struct {
	req   *HttpInReq
	res   *HttpOutRes
	reply *Pipe
}

func NewHttpServeJob(req *HttpInReq, res *HttpOutRes, reply *Pipe) *HttpServeJob {
	return &HttpServeJob{
		req:   req,
		res:   res,
		reply: reply,
	}
}

func (j *HttpServeJob) FinalType() *Type {
	return httpServeJobType
}

func registerHttpServeJobMembers(t *Type) {
	t.registerProperty("req", "HttpInReq",
		func(inst Any) Any { return inst.(*HttpServeJob).req },
		nil,
	)
	t.registerProperty("res", "HttpOutRes",
		func(inst Any) Any { return inst.(*HttpServeJob).res },
		nil,
	)
	t.registerProperty("reply", "Pipe",
		func(inst Any) Any { return inst.(*HttpServeJob).reply },
		nil,
	)
}

func CreateRegisterHttpServeJobType() {
	t := &Type{
		name:            "HttpServeJob",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[HttpServeJob]("HttpServeJob"),
		restrict:        SR_HttpListen,
	}
	registerHttpServeJobMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	httpServeJobType = t
}

var httpServeJobType *Type

func init() {
	CreateRegisterHttpServeJobType()
}

var _ Typed = (*HttpServeJob)(nil)
