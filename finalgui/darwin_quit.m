//go:build darwin

#import <AppKit/AppKit.h>

extern void goNativeQuit(void);

@interface FinalNativeQuitHandler : NSObject
@end

@implementation FinalNativeQuitHandler
+ (void)quit:(id)sender {
	goNativeQuit();
}
@end

void installNativeQuitHookC(void) {
	NSMenu *mainMenu = [NSApp mainMenu];
	if (mainMenu == nil || [mainMenu numberOfItems] < 1) {
		return;
	}
	NSMenu *appMenu = [[mainMenu itemAtIndex:0] submenu];
	if (appMenu == nil) {
		return;
	}
	for (NSMenuItem *item in [appMenu itemArray]) {
		if ([[item keyEquivalent] isEqualToString:@"q"]) {
			[item setTarget:[FinalNativeQuitHandler class]];
			[item setAction:@selector(quit:)];
			return;
		}
	}
}
