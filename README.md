# go-env-flags

![Build Status](https://github.com/TubbyStubby/go-env-flags/actions/workflows/build.yml/badge.svg)
[![Go Reference](https://pkg.go.dev/badge/github.com/TubbyStubby/go-env-flags.svg)](https://pkg.go.dev/github.com/TubbyStubby/go-env-flags)


## Install

```sh
go get github.com/TubbyStubby/go-env-flags
```

## Usage

Package env provides an `env` struct field tag to marshal and unmarshal environment variables and command line flags.

Values can be provided via environment variables or command line flags. Flags take precedence over environment variables.

```go
type Environment struct {
	// Creates flag -home
	Home string `env:"HOME"`

	Jenkins struct {
		// Creates flag -build-id
		BuildId *string `env:"BUILD_ID"`

		// Creates flag -build-number
		BuildNumber int `env:"BUILD_NUMBER"`

		// Creates flag -ci
		Ci bool `env:"CI"`
	}

	Node struct {
		// Multiple env vars mapping to same field
		// Creates flag -npm-config-cache
		ConfigCache *string `env:"npm_config_cache,NPM_CONFIG_CACHE"`
	}

	// Custom flag name -my-flag takes precedence over env generated flag name
	CustomFlag string `env:"SOME_ENV,flag=my-flag"`

	Duration      time.Duration `env:"TYPE_DURATION"`
	DefaultValue  string        `env:"MISSING_VAR,default=default_value"`
	RequiredValue string        `env:"IM_REQUIRED,required=true"`
	ArrayValue    []string      `env:"ARRAY_VALUE,separator=|,default=value1|value2|value3"`
}

func main() {
	var environment Environment

	// Will result in an error if `IM_REQUIRED` is not set in the environment or
	// via flag `-im-required`.
	flags, es, err := env.UnmarshalFromEnviron(&environment)
	if err != nil {
		log.Fatal(err)
	}
}
```

## Auto Flag Name Generation

Flag names are automatically generated from environment variable names using the following rules:

1. All characters are converted to lowercase
2. Non alphanumeric characters are converted to hyphens (-)
3. When transitioning from lowercase to uppercase in the original env name, a hyphen is inserted

Examples:

```go
// Environment Key -> Flag Name
"HOME"              -> "-home"
"BUILD_ID"          -> "-build-id"
"NPM_CONFIG_CACHE"  -> "-npm-config-cache"
"myEnvVar"          -> "-my-env-var"
"API_URLv2"         -> "-api-urlv2"
```

You can provide custom flag name using the `flag` tag option:

```go
type Config struct {
    // Creates flag -custom-name
    Value string `env:"MY_ENV_VAR,flag=custom-name"`
}
```

## Multiple Environment Variables and Flags

The package supports mapping multiple environment variables to a single field. The first environment variable or flag with a value is used.

### Priority

1. Command line flags take precedence over environment variables
2. For multiple environment variables, the first one found with a value is used
3. If no environment variable has a value, the default value is used (if specified)

### Examples:

```go
type Config struct {
    Cache string `env:"npm_config_cache,NPM_CONFIG_CACHE,flag=npm-cache"`
}
```

### Resolution Order:

For the above example:

1. `-npm-cache` custom flag value (if provided)
2. `-npm-config-cache` auto generated flag value (if provided)
3. `npm_config_cache` environment variable (if set)
4. `NPM_CONFIG_CACHE` environment variable (if set)
5. Default value (if specified)

## Flag Descriptions

You can add descriptions to flags that appear in the help output using the `desc` tag option.
Descriptions include information about environment variables, default values, and whether the field is required.

### Basic Description

```go
type Config struct {
    // Basic description
    Port int `env:"PORT,desc=The port number for the server"`

    // Description with default value
    Host string `env:"HOST,default=localhost,desc=The host address to bind to"`

    // Required field with description
    APIKey string `env:"API_KEY,required=true,desc=API key for authentication"`
}
```

Help output:
```
  -port
        The port number for the server. Environment: PORT
  -host
        The host address to bind to. Environment: HOST. Default: localhost
  -api-key
        API key for authentication. Environment: API_KEY. Required: true
```

## Custom Marshaler/Unmarshaler

NOTE: this is only available for environment variables.

[Documentation can be found on upstream.](https://github.com/Netflix/go-env/tree/6b7f89893152c6fd09ac70c4bc7d7d7ed7df5aba?tab=readme-ov-file#custom-marshalerunmarshaler)
