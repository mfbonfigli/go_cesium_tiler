# Go Cesium Tiler - Development Setup Guide

```
                                             _   _ _
  __ _  ___   ___ ___  ___(_)_   _ _ __ ___ | |_(_) | ___ _ __
 / _  |/ _ \ / __/ _ \/ __| | | | | '_   _ \| __| | |/ _ \ '__|
| (_| | (_) | (_|  __/\__ \ | |_| | | | | | | |_| | |  __/ |
 \__, |\___/ \___\___||___/_|\__,_|_| |_| |_|\__|_|_|\___|_|
  __| | A Cesium Point Cloud tile generator written in golang
 |___/ 
```

## Introduction
With the release of version 2, in particular from version 2.0.0-gamma, gocesiumtiler uses the Proj v9.5+ library to perform coordinate conversions. As a result, building the executable is more complex due to the need for `cgo` compiling, but also the need of statically building and linking Proj with the go executable.

The build environment thus needs to be properly setup to enable the builds.

## Reproducible builds powered by Docker
In order to streamline the build steps, reproducible builds are achieved through a Dockerfile. The repository contains three files:
- `Dockerfile`: Contains all the build steps needed to build and test gocesiumtiler in both Windows and Linux.
- `build.sh`: Kickstarts the docker build process injecting the right arguments.
- `build.ps1`: Powershell scripts that works as `build.sh` but meant to be used with powershell under a windows environment.

The Dockerfile is organized as a multi-stage build.
1. A base image is prepared, containing essential build tools.
2. A linux build image is created, where the required dependencies to build under linux are pulled and compiled. Proj is rebuilt from the sources with static linking  targeting the Linux OS and then gocesiumtiler is compiled linking it to the static version of the Proj library.
3. Another linux build image is created, this one pulling the dependencies needed to cross-compile the gocesiumtiler executable under Windows. Proj is rebuilt again, this time targeting the Windows x86-64 architecture.
4. The build artifacts are copied in a final scratch image where they are ready to be copied out to the host.

## Local Development environment setup
Two approaches are possible, using docker to run builds in a reproducible environment or building it locally setting up the local machine.

Note that all the options mentioned statically link Proj to the final build executable, this ensures the system builds a single highly portable binary that embeds all required dependencies. 

### Option 1. Install docker and use `build.sh` or `build.ps1` to build the code and run the tests. 

The pro here is that this should work in any environment and ensures that the final executable build is reproducible. It recommended to always run a "dockerized" build before creating a PR.

To build just run in a powershell console:
```
./build.ps1
```

### Option 2. Setup the local machine for local development without Docker.

This approach is more challenging and the steps are different depending whether you are development under Windows or Linux. In general the environment needs to be set up mimicking the steps described in the `Dockerfile`. 

Since the dockerfile has been written designed to run builds in Linux, that is also documenting how to setup a linux dev environment. In the following we will focus mostly instead on the steps to setup a dev environment under Windows.

### **Setup a Windows development environment**
The commands in the following are supposed to be executed from a powershell console.

**1. Install golang**

**2. Install msys2**

Msys2 will be used to install build tools like the Mingw64 compiler, CMake and Pkgconfig.
Install it following the instructions here https://www.msys2.org/#installation or using Chocolatey.

Make sure the `bin` folder is available on the `Path`. The `Path` should contain 
the following folder (please adapt them according to where your Msys2 installation is located):
- `C:\msys64\usr\bin`

**3. Use Msys2 package manager pacman to install the required build tools**

```
pacman -S --noconfirm mingw-w64-x86_64-pkgconf
pacman -S --noconfirm mingw-w64-x86_64-gcc
pacman -S --noconfirm mingw-w64-x86_64-cmake
pacman -S --noconfirm mingw-w64-x86_64-sqlite3
```

Make sure that mingw64 executable is on the `Path` and if not add it manually, i.e. add :
```
C:\msys64\mingw64\bin
```
to the path.

**4. Install `Vcpkg`, used to manage the dependencies**

[vcpkg](https://vcpkg.io/en/) is a free C/C++ package manager that will greatly simplify the build setup. To install typically just clone the git repository in some folder, eg:
```
git clone https://github.com/Microsoft/vcpkg.git "C:\vcpkg"
```

And then run:
```
cd C:\vcpkg
bootstrap-vcpkg.bat -disableMetrics
```

**5. Set the default Triplets for Vcpkg by creating / setting these environment variables**

- `VCPKG_DEFAULT_TRIPLET`=`x64-mingw-static`
- `VCPKG_DEFAULT_HOST_TRIPLET` = `x64-mingw-static`

**6. Install the dependencies needed to build the Proj library (sqlite3 and tiff) via `vcpkg.exe`**

```
vcpkg.exe install sqlite3[core,tool] zlib --triplet=x64-mingw-static
```

Note, we are building a statically linked version of these libraries.

**7. Clone Proj in some folder, eg `C:\proj`**

```
git clone https://github.com/OSGeo/PROJ.git  "C:\proj"
```

If you want a specific version of Proj instead download and uncompress the archive from
```
https://download.osgeo.org/proj/$PROJ_VERSION.tar.gz
```

where `$PROJ_VERSION` is the name of the version you want, e.g. `proj-9.5.0`

**8. Build PROJ with static linking**

```
cd C:\proj
mkdir build
cd build
cmake -DCMAKE_TOOLCHAIN_FILE=C:\vcpkg\scripts\buildsystems\vcpkg.cmake `
    -DVCPKG_TARGET_TRIPLET=x64-mingw-static `
    -DCMAKE_C_COMPILER=x86_64-w64-mingw32-gcc `
    -DCMAKE_CXX_COMPILER=x86_64-w64-mingw32-g++ `
    -DCMAKE_INSTALL_PREFIX=/usr/local/ `
    -DCMAKE_BUILD_TYPE=Release `
    -DBUILD_APPS=OFF `
    -DBUILD_SHARED_LIBS=OFF `
    -DENABLE_CURL=OFF `
    -DENABLE_TIFF=ON `
    -DEMBED_PROJ_DATA_PATH=OFF `
    -DBUILD_TESTING=OFF .. 
cmake --build . --config Release -j $(nproc)
cmake --build . --target install -j $(nproc)
```

Make sure `which cmake` points to the Msys2 - Mingw64 installation folder. If `$(nproc)` doesn't work just replace it with the number of CPUs on your system.

---

**Note: Steps 1-8 are only required to be executed once, or when you want to upgrade the Proj version or one of its dependencies.**

---

**9. Build gocesiumtiler**

This could require some adaptations, and depending on your development configuration you might have to tune parameters like the linker search path or the `PKG_CONFIG_PATH`.

The following is a general guide of a configuration that could work: 
```
$env:PKG_CONFIG_PATH="C:\usr\local\lib\pkgconfig;C:\vcpkg\installed\x64-mingw-static\lib\pkgconfig"; `
$env:CC="x86_64-w64-mingw32-gcc"; `
$env:CGO_ENABLED=1; `
$env:CGO_LDFLAGS="-L/vcpkg/installed/x64-mingw-static/lib -g -O2 -static -lstdc++ -lsqlite3 -ltiff -lzlib -ljpeg -llzma -lm"; `
go build -o ./bin/gocesiumtiler.exe ./cmd/main.go
```

To run the tests similarly run:
```
$env:PKG_CONFIG_PATH="C:\usr\local\lib\pkgconfig;C:\vcpkg\installed\x64-mingw-static\lib\pkgconfig"; `
$env:CC="x86_64-w64-mingw32-gcc"; `
$env:CGO_ENABLED=1; `
$env:CGO_LDFLAGS="-L/vcpkg/installed/x64-mingw-static/lib -g -O2 -static -lstdc++ -lsqlite3 -ltiff -lzlib -ljpeg -llzma -lm"; `
go test -v ./...
```

The environment variables `PKG_CONFIG_PATH`, `CC`, `CGO_ENABLED`, `CGO_LDFLAGS` could also be stored permanently in the environment configuration so that the build and test commands become trivial. 

```
go build -o ./bin/gocesiumtiler.exe ./cmd/main.go
```

and

```
go test v ./...
```

### **Setup a Linux development environment**

Please refer to the Dockerfile where the commands to setup a dev environment for Ubuntu-based development environment are listed in detail.