### Logging in Cloudinary SDK
The default logger in Cloudinary Go SDK is `go log`.

You can use any log library by overwriting the standard SDK logging functions.

#### Using logrus with the SDK
```go
package main

import (
	"github.com/cloudinary/cloudinary-go"
	"github.com/sirupsen/logrus"
	"log"
)

func main() {
	// Start by creating a new instance of Cloudinary using CLOUDINARY_URL environment variable.
	// Alternatively you can use cloudinary.NewFromParams() or cloudinary.NewFromURL().
	var cld, err = cloudinary.New()
	if err != nil {
		log.Fatalf("Failed to intialize Cloudinary, %v", err)
	}

	// Initialize your logger somewhere in your code.
	// Set cloudinary.Logger.Writer with logrus instance
	var logger = logrus.New()
	cld.Logger.Writer = logger.WithField("source", "cloudinary")
}
```

#### Using Zap with the SDK
```go
package main

import (
	"github.com/cloudinary/cloudinary-go"
	"go.uber.org/zap"
	"log"
)

func main() {
	// Start by creating a new instance of Cloudinary using CLOUDINARY_URL environment variable.
	// Alternatively you can use cloudinary.NewFromParams() or cloudinary.NewFromURL().
	var cld, err = cloudinary.New()
	if err != nil {
		log.Fatalf("Failed to intialize Cloudinary, %v", err)
	}

	// Initialize your logger somewhere in your code.
	// Set cloudinary.Logger.Writer with zap.SugaredLogger instance
	var zapLogger, _ = zap.NewDevelopment()
	cld.Logger.Writer = zapLogger.Sugar().With("source", "cloudinary")
}
```

#### Logging level

You can change logging level with the `Logger.SetLevel()` function.

Possible values:
- `logger.NONE`  - disabling logging from the SDK
- `logger.ERROR` - enable logging only for error messages
- `logger.DEBUG` - enable debug logs
