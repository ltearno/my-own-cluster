(module
  (type (;0;) (func (result i32)))
  (type (;1;) (func (param i32 i32 i32) (result i32)))
  (type (;2;) (func))
  (type (;3;) (func (param i32 i32)))
  (type (;4;) (func (param i32)))
  (import "env" "io_get_stdout" (func $io_get_stdout (type 0)))
  (import "env" "resource_write" (func $resource_write (type 1)))
  (func $__wasm_call_ctors (type 2))
  (func $_start (type 2)
    call $runtime.initAll)
  (func $runtime.initAll (type 2)
    (local i32 i32)
    memory.size
    i32.const 16
    i32.shl
    i32.const 65618
    i32.sub
    i32.const 6
    i32.shr_u
    local.set 0
    i32.const 0
    local.set 1
    block  ;; label = @1
      block  ;; label = @2
        loop  ;; label = @3
          local.get 0
          local.get 1
          i32.eq
          br_if 1 (;@2;)
          i32.const 0
          i32.const 65618
          i32.sub
          local.get 1
          i32.eq
          br_if 2 (;@1;)
          local.get 1
          i32.const 65618
          i32.add
          i32.const 0
          i32.store8
          local.get 1
          i32.const 1
          i32.add
          local.set 1
          br 0 (;@3;)
        end
      end
      i32.const 0
      call $io_get_stdout
      i32.store offset=65536
      return
    end
    call $runtime.nilPanic
    unreachable)
  (func $cwa_main (type 2)
    call $runtime.initAll
    i32.const 65606
    i32.const 12
    call $runtime.printstring
    call $runtime.printnl)
  (func $runtime.printstring (type 3) (param i32 i32)
    (local i32)
    i32.const 0
    local.set 2
    block  ;; label = @1
      loop  ;; label = @2
        local.get 2
        local.get 1
        i32.ge_s
        br_if 1 (;@1;)
        local.get 0
        local.get 2
        i32.add
        i32.load8_u
        call $runtime.putchar
        local.get 2
        i32.const 1
        i32.add
        local.set 2
        br 0 (;@2;)
      end
    end)
  (func $runtime.printnl (type 2)
    i32.const 13
    call $runtime.putchar
    i32.const 10
    call $runtime.putchar)
  (func $runtime.getFuncPtr (type 2)
    (local i32)
    global.get 0
    i32.const 16
    i32.sub
    local.tee 0
    global.set 0
    local.get 0
    i64.const 0
    i64.store offset=8
    call $runtime.nilPanic
    unreachable)
  (func $runtime.nilPanic (type 2)
    call $runtime.runtimePanic
    unreachable)
  (func $memset (type 1) (param i32 i32 i32) (result i32)
    (local i32 i32)
    i32.const 0
    local.set 3
    i32.const 0
    local.get 0
    i32.sub
    local.set 4
    block  ;; label = @1
      block  ;; label = @2
        loop  ;; label = @3
          local.get 2
          local.get 3
          i32.eq
          br_if 1 (;@2;)
          local.get 4
          local.get 3
          i32.eq
          br_if 2 (;@1;)
          local.get 0
          local.get 3
          i32.add
          local.get 1
          i32.store8
          local.get 3
          i32.const 1
          i32.add
          local.set 3
          br 0 (;@3;)
        end
      end
      local.get 0
      return
    end
    call $runtime.nilPanic
    unreachable)
  (func $runtime.runtimePanic (type 2)
    i32.const 65584
    i32.const 22
    call $runtime.printstring
    i32.const 65552
    i32.const 23
    call $runtime.printstring
    call $runtime.printnl
    unreachable
    unreachable)
  (func $runtime.putchar (type 4) (param i32)
    (local i32)
    global.get 0
    i32.const 16
    i32.sub
    local.tee 1
    global.set 0
    local.get 1
    i32.const 0
    i32.store offset=12
    local.get 1
    local.get 0
    i32.store8 offset=12
    i32.const 0
    i32.load offset=65536
    local.get 1
    i32.const 12
    i32.add
    i32.const 1
    call $resource_write
    drop
    local.get 1
    i32.const 16
    i32.add
    global.set 0)
  (func $resume (type 2)
    call $runtime.getFuncPtr
    unreachable)
  (table (;0;) 1 1 anyfunc)
  (memory (;0;) 16)
  (global (;0;) (mut i32) (i32.const 65536))
  (global (;1;) i32 (i32.const 65618))
  (global (;2;) i32 (i32.const 65536))
  (global (;3;) i32 (i32.const 65618))
  (global (;4;) i32 (i32.const 1024))
  (export "memory" (memory 0))
  (export "__wasm_call_ctors" (func $__wasm_call_ctors))
  (export "_start" (func $_start))
  (export "cwa_main" (func $cwa_main))
  (export "__heap_base" (global 1))
  (export "memset" (func $memset))
  (export "resume" (func $resume))
  (export "__dso_handle" (global 2))
  (export "__data_end" (global 3))
  (export "__global_base" (global 4))
  (data (;0;) (i32.const 65536) "\00\00\00\00")
  (data (;1;) (i32.const 65552) "nil pointer dereference\00\00\00\00\00\00\00\00\00panic: runtime error: Hello world!"))
