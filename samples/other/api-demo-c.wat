(module
  (type (;0;) (func (result i32)))
  (type (;1;) (func (param i32 i32) (result i32)))
  (type (;2;) (func (param i32) (result i32)))
  (type (;3;) (func (param i32 i32 i32) (result i32)))
  (import "my-own-cluster" "test" (func (;0;) (type 0)))
  (import "my-own-cluster" "register_buffer" (func (;1;) (type 1)))
  (import "my-own-cluster" "get_buffer_size" (func (;2;) (type 2)))
  (import "my-own-cluster" "get_buffer" (func (;3;) (type 3)))
  (import "my-own-cluster" "free_buffer" (func (;4;) (type 2)))
  (func (;5;) (type 1) (param i32 i32) (result i32)
    (local i32 i32 i32 i32)
    global.get 0
    i32.const 32
    i32.sub
    local.tee 2
    global.set 0
    call 0
    drop
    i32.const 0
    local.set 3
    local.get 2
    i32.const 28
    i32.add
    i32.const 0
    i32.load16_u offset=1028 align=1
    i32.store16
    local.get 2
    i32.const 0
    i32.load offset=1024 align=1
    i32.store offset=24
    i32.const 100
    local.set 4
    block  ;; label = @1
      local.get 2
      i32.const 24
      i32.add
      i32.const 5
      call 1
      local.tee 5
      call 2
      i32.const 5
      i32.ne
      br_if 0 (;@1;)
      local.get 5
      local.get 2
      i32.const 14
      i32.add
      i32.const 5
      call 3
      drop
      loop  ;; label = @2
        block  ;; label = @3
          local.get 2
          i32.const 24
          i32.add
          local.get 3
          i32.add
          i32.load8_u
          local.get 2
          i32.const 14
          i32.add
          local.get 3
          i32.add
          i32.load8_u
          i32.eq
          br_if 0 (;@3;)
          local.get 3
          local.set 4
          br 2 (;@1;)
        end
        local.get 3
        i32.const 1
        i32.add
        local.tee 3
        i32.const 5
        i32.ne
        br_if 0 (;@2;)
      end
      local.get 5
      call 4
      drop
      i32.const 700
      local.set 4
    end
    local.get 2
    i32.const 32
    i32.add
    global.set 0
    local.get 4)
  (func (;6;) (type 2) (param i32) (result i32)
    local.get 0
    call 2)
  (table (;0;) 1 1 funcref)
  (memory (;0;) 1)
  (global (;0;) (mut i32) (i32.const 5136))
  (export "memory" (memory 0))
  (export "_start" (func 5))
  (export "get_size_of_passed_buffer" (func 6))
  (data (;0;) (i32.const 1024) "hello\00"))
