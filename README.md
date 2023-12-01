# unifont

[![Go Reference](https://pkg.go.dev/badge/github.com/steelseries/unifont.svg)](https://pkg.go.dev/github.com/steelseries/unifont)

A go library for using [GNU Unifont](https://unifoundry.com/unifont/index.html) with Go. Implements
the `golang.org/x/image/font.Face` interface for using with that package.

## Use

    import "github.com/steelseries/unifont"

    ...

    uf, err := unifont.ParseHexFile("unifont.hex")
    if err != nil {
        panic(err)
    }

    face, err := unifont.NewFace(uf, 1)
    if err != nil {
        panic(err)
    }

    // use face like any other font.Face, with font.Drawer/etc.
