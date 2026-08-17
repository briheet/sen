(* Senbon function tracer. Linked into the instrumented target program.
   __p_incr is called on every function invocation (inserted by instrument.ml)
   and emits a Runtime_events custom int span carrying the function's id, so an
   external consumer (senbon) sees which functions execute and when. *)

type Runtime_events.User.tag += SenbonFn

let tag =
  Runtime_events.User.register "senbon.fn" SenbonFn Runtime_events.Type.int

let __p_incr (id : int) (f : unit -> 'a) : 'a =
  Runtime_events.User.write tag id;
  f ()

let () = Runtime_events.start ()
