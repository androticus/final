# Installation

Final is available for Windows, Mac, Linux, and Pi. This version, v0.0.10 provides support for:
* macOS Apple Silicon (arm64);
* macOS Intel (amd64);
* Windows (amd64); 
* Linux (amd64/arm64);
* Pi OS 64, for Pi 5 and Pi Zero 2W;
* Pi OS 32, for Pi Zero 2W (TUI version only)

Final is easy to install, and easy to remove, should you wish to do so. It is provided as a simple .zip file, for each supported platform/cpu combination.

## Install instructions

1. Go to the downloads page, and download the .zip for your platform/cpu.
2. Move the .zip to your home folder.
3. Double-click the .zip in your UI, or cd to your home folder and run the unzip program for your OS.
4. This should create a folder `Final` with the `final` and `final-gui` binaries (`final.exe` / `final-gui.exe` on Windows) in the root, and Final's bootstrap files, and example programs.
5. Launch that file in your UI.
6. (optional) Select one of the examples in Final/examples and open it in your favorite editor; associate the .fnl extension with your editor.

## Troubleshooting

Prefer launching from inside the unzipped `Final/` folder. From a shell, `cd` into `Final/` then run `./final` or `./final-gui` (`final.exe` or `final-gui.exe` on Windows).

## Uninstall instructions

1. Delete the folder `Final` in your home folder. That's it!
