# Minipaint

Final provides a "lite" version of Painter call Minipaint, based on the AdaFruit GSV drawing library for C. It uses a BlitSrc (Mat+Palette) as its drawing surface. It is primarily intended for drawing small bit-mapped text and images intended for small IoT/Maker types of displays.

The API of Minipaint is similar to Painter. It also contains a 5x7 bitmapped font which can be scaled when drawing. See the examples/pi folder for the e-ink and oled demos, which use this utility class.
