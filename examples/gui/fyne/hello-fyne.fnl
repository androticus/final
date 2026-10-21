forget: reset-hello-fyne

requiresFyne
// `examples/gui/fyne/hello-fyne.fnl` compileFromFile
`
Hello World using Fyne.

Shows how to:
1. create a Window from the App
2. create a Label
3. set the Window content to the label
4. show the Window
5. after a delay, update the label text

Try:

hello-fyne

Then wait 4 seconds.
`
const: hello-fyne-blurb


: hello-fyne
  App.app "Hello Fyne Application" Window.new^ >ts
  
  "Hello from Fyne!" Label.new^ dup >xs
  t:.setContent^
    
  t:.show^
  
  4000 delayMs
  "See ya later Fynagator!" x:.setText^
  
  x:drop t:drop 
;

hello-fyne-blurb $ nl
