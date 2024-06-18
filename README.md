
```

 _____                           _                  _                    ___   
(  _  )               _         ( )                ( )                  (  _`\ 
| ( ) |   _ _  _ _   (_) ______ | |/')  _ __   _ _ | |/')    __    ___  | | ) |
| | | | /'_` )( '_`\ | |(______)| , <  ( '__)/'_` )| , <   /'__`\/' _ `\| | | )
| (_) |( (_| || (_) )| |        | |\`\ | |  ( (_| || |\`\ (  ___/| ( ) || |_) |
(_____)`\__,_)| ,__/'(_)        (_) (_)(_)  `\__,_)(_) (_)`\____)(_) (_)(____/'
              | |                                                              
              (_)                                                              
                                                            
```



`oapi-krakend` is a command-line tool to convert an OpenAPI specification into KrakenD configurations.

## Installation

1. Install Go version `go1.22.3` or higher.
2. Run the following command to build the project:
   
    ```sh
    go build .
    ```

## Usage

### Required Arguments

- `--spec` : The URL or filepath of the OpenAPI specification.

### Optional Arguments

- `--includes` : A comma-separated list of URI resources to include from the OpenAPI specification paths.
- `--excludes` : A comma-separated list of URI resources to exclude from the OpenAPI specification paths.
- `--outputDir`: The directory where the generated scaffold of KrakendD [Flexible Configuration](https://www.krakend.io/docs/configuration/flexible-config/) templates will be output.

## Examples

### Basic Usage

To create KrakenD configurations from an OpenAPI specification:

```sh
./oapi-krakend --spec="https://api.utiligize.com/openapi.json"
```

### Using the `--includes` Option

To include specific URI resources:

```sh
./oapi-krakend --spec="https://api.utiligize.com/openapi.json" --includes="assets,tasks"
```

### Using the `--excludes` Option

To exclude specific URI resources:

```sh
./oapi-krakend --spec="https://api.utiligize.com/openapi.json" --excludes="der_types"
```


## Output

```plaintext
cfg-krakend/
├── templates/
│   ├── endpoints-1.tmpl
│   ├── endpoints-2.tmpl
│   └── endpoints-N.tmpl
└── krakend.tmpl