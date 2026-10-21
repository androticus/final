# Fyne and final-gui

Fyne is a library of Graphical Interface components, and helper classes, similar in design to many other such libraries. You compose an interface with a combination of layout components, and individual widgets.

There are two central components to any interface you might build:
* `App` which is the top-level Fyne object; you access this in Final via `App.app`.
* `Window` which is the root object that contains all your other elements. Individual widgets can only be shown once added to a Window, via its `setContent` method, which must either be a Layout or Widget.

There are numerous examples of how to use Fyne in `examples/gui/fyne`. Here is a simple "Hello Fyne!" demo that displays the venerable message in a Label widget, then updates the text four seconds later. It shows how easy it is to set up and display a Fyne Window.

```
: hello-fyne
  App.app "Hello Fyne Application" Window.new^ >ts
  
  "Hello from Fyne!" Label.new^ dup >xs
  t:.setContent^
    
  t:.show^
  
  4000 delayMs
  "See ya later Fynagator!" x:.setText^
  
  x:drop t:drop 
;
```

## Layouts

A `Layout` is a component that helps organize the widgets it contains, so you don't have to try manually setting the position of the widgets, which can be tedious, and fragile when you make changes.

Fyne has simple layouts, like a horizontal and vertical Box, which layout the widgets in those orders. It has a Form layout, for aligning labels with widgets in a grid. It has a sophisticated Border layout, which can place top/bottom/left/right/ and a central widget on the screen.

## Widgets

Final supports a variety of Fyne's widget types, such as Label, Entry, Button, and so on. These are covered in detail in the next chapter.

## Menus

Fyne supports menu bars, menus, and menu items. The Fyne app `examples/gui/fyne/widgets-fyne.fnl` shows how to add a menu bar to your window.

## Events

Fyne provides change notification, such as Tapping a button, typing in an Entry field, and selecting a menu item via Events, the basics of which were covered in [Events and EventLoop](../events.md). You create an event handler in a Final method on a TypedObject instance, then use a MethodRef to obtain a reference to that method, which, along with an index, you pass to a Fyne api that requires or accepts an Event Handler. When the event is triggered, your handler Method runs, with your TypedObject instance on the O Stack, where you can read/write its properties, or call other methods or helper words.

## Preferences

Fyne manages a local storage location into which you can save and fetch preferences associated with your app by using the Prefs class object.