# Level Change Events (1)

Many control situations involve doing something when the level of an input changes, either in one direction, or on both edges. For example, a button typically causes its action when the level changes from unpressed to pressed.

To detect level changes in a loop-style controller, you create a variable that holds the previous state of the pin, then in the loop, check for the current state being different. Then you typically set a flag that is picked up later in the loop, or call some action code immediately.

Final provides a second approach, which is to raise an event and call a TypedObject method you provide when the level changes. You can independently set an event handler for the pin going low and going high. We discuss how to use and handle events later in the controller section [Event Handlers](ctrl-events.md).
