(* Static analyzer for an OCaml project: emits a JSON callgraph.
   Usage: <binary> <root.ml> <output.json>
   Built against OCaml compiler-libs (parser + AST). *)
let read_file fn =
  let ic = open_in_bin fn in
  let n = in_channel_length ic in
  let s = really_input_string ic n in
  close_in ic; s

let json_string s =
  let b = Buffer.create 64 in
  Buffer.add_char b '"';
  String.iter (fun c ->
    match c with
    | '"' -> Buffer.add_string b "\\\""
    | '\\' -> Buffer.add_string b "\\\\"
    | '\n' -> Buffer.add_string b "\\n"
    | c' -> Buffer.add_char b c') s;
  Buffer.add_char b '"';
  Buffer.contents b

let () =
  if Array.length Sys.argv < 3 then (prerr_endline "usage: analyze <root.ml> <out.json>"; exit 2);
  let root = Sys.argv.(1) in
  let out = Sys.argv.(2) in
  let src = read_file root in
  let buf = Lexing.from_string src in
  buf.Lexing.lex_curr_p <- { buf.Lexing.lex_curr_p with Lexing.pos_fname = root };
  let ast =
    try Parse.implementation buf
    with _ -> prerr_endline ("parse error: " ^ root); exit 1
  in

  (* function id table *)
  let names = Hashtbl.create 32 in
  let next_id = ref 0 in
  let fid name =
    match Hashtbl.find_opt names name with
    | Some i -> i
    | None -> let i = !next_id in incr next_id; Hashtbl.add names name i; i
  in
  (* edges (from,to) dedup *)
  let edges = Hashtbl.create 64 in
  let add_edge f t = if not (Hashtbl.mem edges (f,t)) then Hashtbl.add edges (f,t) () in

  let cur = ref (-1) in
  let module IT = Ast_iterator in
  let iterator =
    { IT.default_iterator with
      expr = (fun self e ->
        (match e.Parsetree.pexp_desc with
         | Parsetree.Pexp_apply (f, _args) ->
             (match f.Parsetree.pexp_desc with
              | Parsetree.Pexp_ident {txt=Longident.Lident n} when !cur >= 0 ->
                  if n <> "" && n.[0] <> '<' then add_edge !cur (fid n)
              | _ -> ())
         | _ -> ());
        IT.default_iterator.expr self e);
      value_binding = (fun self vb ->
        let name = match vb.Parsetree.pvb_pat.Parsetree.ppat_desc with
          | Parsetree.Ppat_var {txt} -> txt
          | _ -> "" in
        let prev = !cur in
        if name <> "" then cur := fid name;
        IT.default_iterator.value_binding self vb;
        cur := prev);
    }
  in
  iterator.structure iterator ast;

  (* build id -> name array *)
  let count = !next_id in
  let arr = Array.make count "" in
  Hashtbl.iter (fun name id -> if id >= 0 && id < count then arr.(id) <- name) names;
  let is_operator n =
    n <> "" && (n.[0] = '<' || n.[0] = '+' || n.[0] = '-' || n.[0] = '*' || n.[0] = '/' 
               || n.[0] = '=' || n.[0] = '!' || n.[0] = '>') in
  (* filter edges to user-defined non-operator callees *)
  let user_edges = Hashtbl.create 64 in
  Hashtbl.iter (fun (f,t) () ->
    if t >= 0 && t < count && not (is_operator arr.(t)) then
      Hashtbl.replace user_edges (f,t) ()) edges;

  (* JSON *)
  let ob = Buffer.create 512 in
  Buffer.add_string ob "{";
  Buffer.add_string ob "\"entry\":";
  Buffer.add_string ob (json_string root);
  Buffer.add_string ob ",\"functions\":[";
  Array.iteri (fun i nm -> if i > 0 then Buffer.add_string ob ","; Buffer.add_string ob (json_string nm)) arr;
  Buffer.add_string ob "],\"edges\":[";
  let first = ref true in
  Hashtbl.iter (fun (f,t) () ->
    if not !first then Buffer.add_string ob ",";
    first := false;
    Printf.bprintf ob "{\"from\":%d,\"to\":%d}" f t) user_edges;
  Buffer.add_string ob "]}";

  let oc = open_out out in
  output_string oc (Buffer.contents ob);
  close_out oc
