# msgkit

msgkit is a simple wrapper for gorillas websocket package. It makes it quick and easy to write a websocket server using traditional http style request/message handlers

## Usage

```go
package main

import (
	"log"

	"github.com/tile38/msgkit"
)

func main() {
	// Initialize a msgkit server
	s := msgkit.New("/ws")

	// Bind a response handler to any JSON message that contains a "type" of "ID"
	s.Handle("ID", func(c *msgkit.Context) {
		c.Conn.Send(c.ConnID)
	})

	// Listen for requests on port 8000
	log.Println(s.Listen(":8000"))
}
```

## The Idea

The msgkit payload is a JSON payload containing AT LEAST a message type. Any websocket message with a "type" field will be passed to its respective handler defined in your go code. You can choose to nest payloads within another field in your JSON message or pass fields at the parent level.

Note: The "type" should indicate both your method AND resource if applicable.

#### A request for an account with ID 1234
```
{
    "type": "Account",
    "id:": 1234,
}
```

#### A message with the text "Hey guys!"
```
{
    "type": "Message",
    "text": "Hey guys!"
}
```

#### A request to create an account
```
{
    "type": "Create-Account",
    "account": {
        "name": "Mike",
        "username": "Mike1234",
        "password": "Goforlife997"
    }
}
```

The MIT License (MIT)
=====================

Copyright © 2018 Tile38, LLC

Permission is hereby granted, free of charge, to any person
obtaining a copy of this software and associated documentation
files (the “Software”), to deal in the Software without
restriction, including without limitation the rights to use,
copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the
Software is furnished to do so, subject to the following
conditions:

The above copyright notice and this permission notice shall be
included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES
OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT
HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
OTHER DEALINGS IN THE SOFTWARE.
