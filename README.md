# Go Cesium Tiler
![Build Status](https://codebuild.eu-west-1.amazonaws.com/badges?uuid=eyJlbmNyeXB0ZWREYXRhIjoiTW5nSHpJcGJTQjJEekVQL2JsbUJWeVhmUHNZMFhsRGh6TXdPaGxSSlhHVkFCQ1RUUVM0YjJ6dWkyNmJaQmVKdE94RDVzNEhwbHZaOHJGakpCMkJCUG00PSIsIml2UGFyYW1ldGVyU3BlYyI6Im9COW9CVm9FM3ljQlR4RmQiLCJtYXRlcmlhbFNldFNlcmlhbCI6MX0%3D&branch=v2)
```
                                             _   _ _
  __ _  ___   ___ ___  ___(_)_   _ _ __ ___ | |_(_) | ___ _ __
 / _  |/ _ \ / __/ _ \/ __| | | | | '_   _ \| __| | |/ _ \ '__|
| (_| | (_) | (_|  __/\__ \ | |_| | | | | | | |_| | |  __/ |
 \__, |\___/ \___\___||___/_|\__,_|_| |_| |_|\__|_|_|\___|_|
  __| | A Cesium Point Cloud tile generator written in golang
 |___/ 
```


Go Cesium Tiler is a tool to convert point cloud stored as LAS files to Cesium.js 3D tiles ready to be
streamed, automatically generating the appropriate level of details and including additional information for each point 
such as color, laser intensity and classification.   

## What's new: V2
GoCesiumTiler V2 has been released in preview mode and it introduces several important improvements over V1:
- Greatly reduced memory usage: you can expect a 60%+ reduction of memory consumption.
- Faster: up to 15% faster compared to V1
- Allows reading and merging multiple LAS files in a single 3D Tile output (will load them all up in memory)
- More intuitive fine tuning of the sampling quality and hard safeguards against deeply nested trees or small tiles
- Assets embedded in the binary: no need to deploy the assets folder, works as a single portable binary
- Ready to be used as part of other go packages with an easy to use interface
- Many bugfixes
- Much greater unit test coverage

This release is backward incompatible compared to V1 and also deprecates several options, most of which were not of much interest.
Some of these might be added in future minor updates of V2.
- Refine mode "REPLACE" is currently unavailable in V2
- Recursive options is currently unavailable in V2
- Algorithm cannot be chosen anymore: grid sampling is the only one available and is implicitly used

## Features
Go Cesium Tiler automatically handles coordinate conversion to the format required by Cesium and can also 
convert the elevation measured above the geoid to the elevation above the ellipsoid as by Cesium requirements. 
The tool uses the version 4.9.2 of the well-known Proj.4 library to handle coordinate conversion. The input SRID is
specified by just providing the relative EPSG code, an internal dictionary converts it to the corresponding proj4 
projection string.

Speed is a major concern for this tool, thus it has been chosen to store the data completely in memory. If you don't 
have enough memory the tool will fail, so if you have really big LAS files and not enough RAM it is advised to split 
the LAS in smaller chunks to be processed separately.

Information on point intensity and classification is stored in the output tileset Batch Table under the 
propeties named `INTENSITY` and `CLASSIFICATION`.


## Changelog
##### Version 2.0.0
* Most of the code has been rewritten from the ground up, achieving much faster tiling with lower memory usage
* Allows merging multiple LAS files into a single 3D tile
* New options to fine tune the sampling quality: min points per tile, resolution and max tree depth
* Assets now embedded in the binary
* Much higher unit test coverage
* The tiler library can be used into external golang projects
* The CLI has been rewritten from scratch, many options changed name or were added/removed
* `GOCESIUMTILER_WORKDIR` has been deprecated

## Precompiled Binaries
Along with the source code a prebuilt binary for Windows x64 is provided for each release of the tool in the github page.
Binaries for other systems at the moment are not provided.

## Environment setup and compiling from sources
To get started with development just clone the repository. 

When launching a build with `go build` go modules will retrieve the required dependencies. 

To build the CLI executable you can use:
```
go build -o gocesiumtiler ./cmd/main.go
```

This will create an executable named `gocesiumtiler` in the build folder.

As the project and its dependencies make use of C code, under windows you should also have GCC compiler installed and available
in the PATH environment variable. More information on cgo compiler are available [here](https://github.com/golang/go/wiki/cgo).

Additionally make sure CGO is enabled via `go env CGO_ENABLED`. CGO_ENABLED environment variable should be set to 1.

Under linux you will have to have `gcc` installed. Also make sure go is configured to pass the correct flags to gcc. In particular if you encounter compilation errors similar to `undefined reference to 'sqrt'` it means that it is not linking the standard math libraries. A way to fix this is to add `-lm` to the `CGO_LDFLAGS`environment variable, for example by running `export CGO_LDFLAGS="-g -O2 -lm"`.

To launch the tests use the command
```
go test ./... -v
```

To check the test coverage use:
```
go test -coverprofile cover.out -v  ./... && go tool cover -html=cover.out
```

## Usage

To run just execute the binary tool with the appropriate flags.

There are various algorithms selectable.

To show help run:
```
gocesiumtiler -help
```

### Commands
There are two commands, `file` and `folder`:

* `gocesiumtiler file { flags } myfile.las`: Converts `myfile.las` into a Cesium 3D point cloud using the flags passed in input (see below).
* `gocesiumtiler folder { flags } myfolder`: Finds all LAS files into `myfolder` and convers them into one or more Cesium 3D Point clouds using the flags passed as input (see below).S

### Flags

#### Common flags
These flags are applicable to both the `file` and the `folder` commands
```
   --out value, -o value                  full path of the output folder where to save the resulting Cesium tilesets
   --epsg value, -e value                 EPSG code of the input coordinate system (default: -1)
   --resolution value, -r value           minimum resolution of the 3d tiles, in meters. approximately represets the maximum sampling distance between any two points at the lowest level of detail (default: 20)
   --z-offset value, -z value             z offset to apply to the point, in meters. only use it if the input elevation is referred to the WGS84 ellipsoid or geoid (default: 0)
   --depth value, -d value                maximum depth of the output tree. (default: 10)
   --min-points-per-tile value, -m value  minimum number of points to enforce in each 3D tile (default: 5000)
   --geoid, -g                            set to interpret input points elevation as relative to the Earth geoid (default: false) 
   --8-bit                                set to interpret the input points color as part of a 8bit color space (default: false)  
   --help, -h                             show help
```

#### Folder command flags
These commands are specific to the `folder` command:
```
   --join, -j                             merge the input LAS files in the folder into a single cloud. The LAS files must have the same properties (CRS etc) (default: false)
```

### Usage examples:

#### Example 1
Convert all LAS files in folder `C:\las`, write output tilesets in folder `C:\out`, assume LAS input coordinates expressed 
in EPSG:32633, convert elevation from above the geoid to above the ellipsoid. Use a resolution of 10 meters, create tiles with minimum 1000 points
and enforce a maximum tree depth of 12:

```
gocesiumtiler folder -out C:\out -epsg 32633 -geoid -resolution 10 -min-points-per-tile 1000 -depth 12 C:\las
```

or, using the shorthand notation:
```
gocesiumtiler folder -o C:\out -e 32633 -g -r 10 -m 1000 -d 12 C:\las
```

#### Example 2
Like Example 1 but merge all LAS files in `C:\las` into a single 3D Tileset

```
gocesiumtiler folder -out C:\out -epsg 32633 -geoid -resolution 10 -min-points-per-tile 1000 -depth 12 -join C:\las
```

or, using the shorthand notation:
```
gocesiumtiler folder -o C:\out -e 32633 -g -r 10 -m 1000 -d 12 -j C:\las
```

#### Example 3
Convert a single LAS file at `C:\las\file.las`, write the output tileset in the folder `C:\out`, use the system defaults

```
gocesiumtiler file -out C:\out -epsg 32633 C:\las\file.las
```
or, using the shorthand notation:

```
gocesiumtiler file -o C:\out -e 32633 C:\las\file.las
```

### Algorithms

The sampling occurs using a hybrid, lazy octree data structure. The algorithm works as follows:
1. All points are stored in a linked tree: this provides efficient list manipulations operations (splitting, adding, removing) and avoids
 dynamic allocations of slices. The coordinates are internally converted to EPSG 4978
2. An octree cell is created. Every cell stores N points (variable). A cell also has a grid spacing property. The root node has a spacing set to the
 provided resolution. 
3. The points are traversed. Each point will fall into one of the cells the space has been divided by the given grid spacing. If the point is 
 the closest one the the center of the cell, it's taken as new closest for that cell, else it's discarded and parked into a list of the Node octant it belongs to.
4. When all points are traversed the points closes to the cells the space has been partitioned in will be the points for the current tree node. All others are parked.
If an octant however has a number of parked points that is less than the min-points-per-node, the parked points for that octant are rolled up to the current node. Similarly 
if the current node has a depth equal to the configured max depth.
5. Whenever the children are retrieved, the previously parked points are used to create child nodes on demand using the same algorithm, lazily.

## Precompiled Binaries
Along with the source code a prebuilt binary for Windows x64 is provided for each release of the tool in the github page.
Binaries for other systems at the moment are not provided.

## Future work and support

Further work needs to be done, such as: 
- Support for 3D Tiles v.1.1 and GLTF
- Upgrading of the Proj4 library to versions newer than 4.9.2
- Adding support for non-metric units for elevations
 
Contributors and their ideas are welcome.

If you have questions you can contact me at <m.federico.bonfigli@gmail.com>

## Versioning

This library uses [SemVer](http://semver.org/) for versioning. 
For the versions available, see the [tags on this repository](https://github.com/mfbonfigli/gocesiumtiler/v2/tags). 

## Credits

**Massimo Federico Bonfigli** -  [Github](https://github.com/mfbonfigli)

## License

This project is licensed under the GNU Lesser GPL v.3 License - see the [LICENSE.md](LICENSE.md) file for details.

The software uses third party code and libraries. Their licenses can be found in
[LICENSE-3RD-PARTIES.md](LICENSE-3RD-PARTIES.md) file.

## Acknowledgments

* Cesium JS library [github](https://github.com/AnalyticalGraphicsInc/cesium)
* TUM-GIS cesium point cloud generator [github](https://github.com/tum-gis/cesium-point-cloud-generator)
* Simon Hege's golang bindings for Proj.4 library [github](https://github.com/xeonx/proj4)
* John Lindsay go library for reading LAS files [lidario](https://github.com/xeonx/proj4)
* Sean Barbeau Java porting of Geotools EarthGravitationalModel code [github](https://github.com/barbeau/earth-gravitational-model)
