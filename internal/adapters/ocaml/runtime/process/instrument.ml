(* Instrument an OCaml source file for per-call function tracing.
   Usage: <binary> <src.ml> <out.ml> <out.json>
   Rewrites each top-level `let f = fun P1..Pn -> body` binding into
       `let f = fun P1..Pn -> __p_incr <id> (fun () -> body)`
   so every invocation of [f] emits a Runtime_events custom span named after its
   id. Emits JSON { "id": "<fn-name>" } mapping span ids to function names. *)
let read_file fn =
  let ic = open_in_bin fn in
  let n = in_channel_length ic in
  let s = really_input_string ic n in
  close_in ic; s

let json_string s =
  let b = Buffer.create 64 in
  Buffer.add_char b '"';
  String.iter (fun c -> match c with
    | '"' -> Buffer.add_string b "\\\""
    | '\\' -> Buffer.add_string b "\\\\"
    | '\n' -> Buffer.add_string b "\\n"
    | c' -> Buffer.add_char b c') s;
  Buffer.add_char b '"';
  Buffer.contents b

let () =
  if Array.length Sys.argv < 4 then (prerr_endline "usage: instrument <src.ml> <out.ml> <out.json>"; exit 2);
  let src = Sys.argv.(1) in
  let outml = Sys.argv.(2) in
  let outjson = Sys.argv.(3) in
  let text = read_file src in
  let buf = Lexing.from_string text in
  buf.Lexing.lex_curr_p <- { buf.Lexing.lex_curr_p with Lexing.pos_fname = src };
  let ast = Parse.implementation buf in

  let next_id = ref 0 in
  let table = Hashtbl.create 32 in
  let fid name =
    match Hashtbl.find_opt table name with
    | Some i -> i
    | None -> let i = !next_id in incr next_id; Hashtbl.add table name i; i
  in

  let module AM = Ast_mapper in
  let module AH = Ast_helper in
  (* longident for Profiler.__p_incr, the per-call span emitter *)
  let prof_lid : Longident.t =
    match Longident.unflatten ["Profiler"; "__p_incr"] with
    | Some lid -> lid
    | None -> assert false
  in
  let mapper =
    { AM.default_mapper with
      value_binding = (fun self vb ->
        let name = match vb.Parsetree.pvb_pat.Parsetree.ppat_desc with
          | Parsetree.Ppat_var {txt} -> txt
          | _ -> "" in
        let body = self.expr self vb.Parsetree.pvb_expr in
        let loc = body.Parsetree.pexp_loc in
        if name = "" then { vb with Parsetree.pvb_expr = body } else
        if name = "__p_incr" then { vb with Parsetree.pvb_expr = body } else
        let id = fid name in
        (* wrap body: __p_incr <id> (fun () -> body) *)
        let idlit = AH.Exp.constant ~loc (AH.Const.int id) in
        let unitpat = AH.Pat.construct ~loc (Location.mkloc (Longident.Lident "()") loc) None in
        let thunk = match body.Parsetree.pexp_desc with
          | Parsetree.Pexp_function (params, constraint_, fbody) ->
              (* rebuild preserving params, wrapping the body *)
              let wrapped_body = match fbody with
                | Parsetree.Pfunction_body e ->
                    let thunk = AH.Exp.function_ ~loc
                        [{ pparam_loc = loc; pparam_desc = Parsetree.Pparam_val (Nolabel, None, unitpat) }] None
                        (Parsetree.Pfunction_body e) in
                    let prof = AH.Exp.apply ~loc
                        (AH.Exp.ident ~loc (Location.mkloc prof_lid loc))
                        [(Nolabel, idlit); (Nolabel, thunk)] in
                    Parsetree.Pfunction_body prof
                | (Parsetree.Pfunction_cases (_l, _loc, _a) as cases) -> cases in
              AH.Exp.function_ ~loc params constraint_ wrapped_body
          | _ ->
              (* non-function body: leave unwrapped (values aren't called) *)
              body in
        { vb with Parsetree.pvb_expr = thunk });
    }
  in
  let ast' = mapper.structure mapper ast in

  let oc = open_out outml in
  let fmt = Format.formatter_of_out_channel oc in
  Pprintast.structure fmt ast';
  Format.pp_print_flush fmt ();
  close_out oc;

  let ob = Buffer.create 128 in
  Buffer.add_string ob "{";
  let count = !next_id in
  let arr = Array.make count "" in
  Hashtbl.iter (fun name id -> if id >= 0 && id < count then arr.(id) <- name) table;
  Array.iteri (fun id name ->
    if id > 0 then Buffer.add_string ob ",";
    Buffer.add_string ob ("\"" ^ string_of_int id ^ "\":" ^ json_string name)) arr;
  Buffer.add_string ob "}";
  let oj = open_out outjson in
  output_string oj (Buffer.contents ob);
  close_out oj
