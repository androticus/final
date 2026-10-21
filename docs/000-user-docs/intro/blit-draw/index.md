# Raster Drawing

Final provides a rich complement of features that support raster-oriented drawing. It includes the following main objects and operations:
* 1, 4, and 8 bit color index matrices
* 1, 4, and 8 bit Palettes
* a full Color matrix
* Pixmap, an image drawing surface
* 1, 4, 8 bit, and full-color (RGBA) blitting operations
* Drawing operations that include lines, rectangles, circles, and many others
* import/export from/to PNG JPG and GIF

All the features above are available in every version of final. The non-gui versions will primarily use these features for image file generation, processing, rendering, and web serving.

The `final-gui` version of Final includes support for the [The Fyne UI Library](..//fyne/index.md), with support for Windows, Canvas, a variety of widgets such as text entry, radio, select, etc. It also includes a variety of layout containers to conveniently lay out an interface. A Pixmap can be set into a CanvasImage widget for direct visualization.

