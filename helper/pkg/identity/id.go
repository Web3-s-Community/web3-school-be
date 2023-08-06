package identity

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

var ServiceID = 0

func init() {
	rand.Seed(time.Now().UnixNano())
	ServiceID = rand.Intn(math.MaxInt)
	fmt.Println("Service ID is: ", ServiceID)
}
