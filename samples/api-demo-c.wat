(module
  (type (;0;) (func (result i32)))
  (type (;1;) (func (param i32 i32) (result i32)))
  (import "my-own-cluster" "test" (func (;0;) (type 0)))
  (func (;1;) (type 1) (param i32 i32) (result i32)
    call 0
    drop
    local.get 1
    local.get 0
    i32.add)
  (table (;0;) 1 1 anyfunc)
  (memory (;0;) 1)
  (global (;0;) (mut i32) (i32.const 5120))
  (export "memory" (memory 0))
  (export "_start" (func 1)))
