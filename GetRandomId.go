package common

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

func GetRandomId() (int, error) {
	return strconv.Atoi(fmt.Sprintf("%06v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(1000000)))
}
