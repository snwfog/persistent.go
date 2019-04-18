package persistent

import (
  "fmt"
  "reflect"
  "strconv"
  "unsafe"

  "github.com/dchest/siphash"
)

const (
  // intSizeBytes is the size in byte of an int or uint value.
  intSizeBytes = strconv.IntSize >> 3

  // generated by splitting the md5 sum of "hashmap"
  sipHashKey1 = 0xdda7806a4847ec61
  sipHashKey2 = 0xb5940c2623a5aabd
)

func getKeyHash(key interface{}) uint64 {
  switch x := key.(type) {
  case string:
    return getStringHash(x)
  case []byte:
    return siphash.Hash(sipHashKey1, sipHashKey2, x)
  case int:
    return getUintptrHash(uintptr(x))
  case int8:
    return getUintptrHash(uintptr(x))
  case int16:
    return getUintptrHash(uintptr(x))
  case int32:
    return getUintptrHash(uintptr(x))
  case int64:
    return getUintptrHash(uintptr(x))
  case uint:
    return getUintptrHash(uintptr(x))
  case uint8:
    return getUintptrHash(uintptr(x))
  case uint16:
    return getUintptrHash(uintptr(x))
  case uint32:
    return getUintptrHash(uintptr(x))
  case uint64:
    return getUintptrHash(uintptr(x))
  case uintptr:
    return getUintptrHash(x)
  }
  panic(fmt.Errorf("unsupported key type %T", key))
}

func getStringHash(s string) uint64 {
  sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
  bh := reflect.SliceHeader{
    Data: sh.Data,
    Len:  sh.Len,
    Cap:  sh.Len,
  }
  buf := *(*[]byte)(unsafe.Pointer(&bh))
  return siphash.Hash(sipHashKey1, sipHashKey2, buf)
}

func getUintptrHash(num uintptr) uint64 {
  bh := reflect.SliceHeader{
    Data: uintptr(unsafe.Pointer(&num)),
    Len:  intSizeBytes,
    Cap:  intSizeBytes,
  }
  buf := *(*[]byte)(unsafe.Pointer(&bh))
  return siphash.Hash(sipHashKey1, sipHashKey2, buf)
}
