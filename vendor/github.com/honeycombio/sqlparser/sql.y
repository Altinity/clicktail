// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

%{
package sqlparser

func SetParseTree(yylex interface{}, stmt Statement) {
  yylex.(*Tokenizer).ParseTree = stmt
}

func SetAllowComments(yylex interface{}, allow bool) {
  yylex.(*Tokenizer).AllowComments = allow
}

func ForceEOF(yylex interface{}) {
  yylex.(*Tokenizer).ForceEOF = true
}

var (
  SHARE = "share"
  MODE  = "mode"
)

%}

%union {
  empty       struct{}
  statement   Statement
  selStmt     SelectStatement
  runes2      [][]rune
  runes       []rune
  run         rune
  str         string
  selectExprs SelectExprs
  selectExpr  SelectExpr
  columns     Columns
  colName     *ColName
  tableExprs  TableExprs
  tableExpr   TableExpr
  smTableExpr SimpleTableExpr
  tableName   *TableName
  indexHints  *IndexHints
  expr        Expr
  boolExpr    BoolExpr
  valExpr     ValExpr
  colTuple    ColTuple
  valExprs    ValExprs
  values      Values
  rowTuple    RowTuple
  subquery    *Subquery
  caseExpr    *CaseExpr
  whens       []*When
  when        *When
  orderBy     OrderBy
  order       *Order
  limit       *Limit
  insRows     InsertRows
  updateExprs UpdateExprs
  updateExpr  *UpdateExpr

/*
for CreateTable
*/
  createTableStmt CreateTable
  columnDefinition *ColumnDefinition
  columnDefinitions ColumnDefinitions
  columnAtts ColumnAtts
}

%token LEX_ERROR
%token <runes> SELECT INSERT UPDATE DELETE FROM WHERE GROUP HAVING ORDER BY LIMIT FOR OFFSET
%token <runes> ALL DISTINCT AS EXISTS IN IS LIKE BETWEEN NULL ASC DESC VALUES INTO DUPLICATE KEY DEFAULT SET LOCK BINARY
%token <runes> ID STRING NUMBER VALUE_ARG LIST_ARG COMMENT
%token <empty> LE GE NE NULL_SAFE_EQUAL
%token <empty> '(' '=' '<' '>' '~'

%token <runes> PRIMARY
%token <runes> UNIQUE
%left <runes> UNION MINUS EXCEPT INTERSECT
%left <empty> ','
%left <runes> JOIN STRAIGHT_JOIN LEFT RIGHT INNER OUTER CROSS NATURAL USE FORCE
%left <runes> ON
%left <runes> OR
%left <runes> AND
%right <runes> NOT
%left <empty> '&' '|' '^'
%left <empty> '+' '-'
%left <empty> '*' '/' '%'
%nonassoc <empty> '.' 
%left <empty> UNARY 
%right <runes> CASE WHEN THEN ELSE
%left <runes> END

// DDL Tokens
%token <runes> CREATE ALTER DROP RENAME ANALYZE ENGINE
%token <runes> TABLE INDEX VIEW TO IGNORE IF USING
%token <runes> SHOW DESCRIBE EXPLAIN

%start any_command

%type <statement> command
%type <selStmt> select_statement
%type <statement> insert_statement update_statement delete_statement set_statement
%type <statement> create_statement alter_statement rename_statement drop_statement
%type <statement> analyze_statement
%type <runes2> comment_opt comment_list
%type <str> union_op
%type <str> distinct_opt
%type <selectExprs> select_expression_list
%type <selectExpr> select_expression
%type <runes> as_lower_opt as_opt
%type <expr> expression
%type <tableExprs> table_expression_list
%type <tableExpr> table_expression
%type <str> join_type
%type <smTableExpr> simple_table_expression
%type <tableName> dml_table_expression
%type <indexHints> index_hint_list
%type <runes2> index_list
%type <boolExpr> where_expression_opt
%type <boolExpr> boolean_expression condition
%type <str> compare
%type <insRows> row_list
%type <valExpr> value value_expression
%type <colTuple> col_tuple
%type <valExprs> value_expression_list
%type <values> tuple_list
%type <rowTuple> row_tuple
%type <runes> keyword_as_func
%type <subquery> subquery
%type <run> unary_operator
%type <colName> column_name
%type <caseExpr> case_expression
%type <whens> when_expression_list
%type <when> when_expression
%type <valExpr> value_expression_opt else_expression_opt
%type <valExprs> group_by_opt
%type <boolExpr> having_opt 
%type <orderBy> order_by_opt order_list
%type <order> order
%type <str> asc_desc_opt
%type <limit> limit_opt
%type <str> lock_opt
%type <columns> column_list_opt column_list
%type <updateExprs> on_dup_opt
%type <updateExprs> update_list
%type <updateExpr> update_expression
%type <empty> exists_opt not_exists_opt ignore_opt non_rename_operation to_opt constraint_opt using_opt
%type <runes> id_or_reserved_word
%type <empty> force_eof

/*
Below are modification to extract primary key
*/
/*
keywords
*/
%token <empty> BIT TINYINT SMALLINT MEDIUMINT INT INTEGER BIGINT REAL DOUBLE FLOAT UNSIGNED ZEROFILL DECIMAL NUMERIC DATE TIME TIMESTAMP DATETIME YEAR
%token <empty> TEXT CHAR VARCHAR MEDIUMTEXT CHARSET

%token <empty> NULLX AUTO_INCREMENT BOOL APPROXNUM INTNUM

%type <str> data_type
%type <columnDefinition> column_definition
%type <columnDefinitions> column_definition_list
%type <statement> create_table_statement
%type <str> length_opt char_type numeric_type unsigned_opt zero_fill_opt key_att int_type decimal_type precision_opt time_type
%type <columnAtts> column_atts



%%

any_command:
  command
  {
    SetParseTree(yylex, $1)
  }

command:
  select_statement
  {
    $$ = $1
  }
| insert_statement
| update_statement
| delete_statement
| set_statement
| create_statement
| alter_statement
| rename_statement
| drop_statement
| analyze_statement

select_statement:
  SELECT comment_opt distinct_opt select_expression_list
  {
    $$ = &Select{Comments: Comments($2), Distinct: $3, SelectExprs: $4}
  }
| SELECT comment_opt distinct_opt select_expression_list FROM table_expression_list where_expression_opt group_by_opt having_opt order_by_opt limit_opt lock_opt
  {
    $$ = &Select{Comments: Comments($2), Distinct: $3, SelectExprs: $4, From: $6, Where: NewWhere(AST_WHERE, $7), GroupBy: GroupBy($8), Having: NewWhere(AST_HAVING, $9), OrderBy: $10, Limit: $11, Lock: $12}
  }
| select_statement union_op select_statement %prec UNION
  {
    $$ = &Union{Type: $2, Left: $1, Right: $3}
  }

insert_statement:
  INSERT comment_opt INTO dml_table_expression column_list_opt row_list on_dup_opt
  {
    $$ = &Insert{Comments: Comments($2), Table: $4, Columns: $5, Rows: $6, OnDup: OnDup($7)}
  }
| INSERT comment_opt INTO dml_table_expression SET update_list on_dup_opt
  {
    cols := make(Columns, 0, len($6))
    vals := make(ValTuple, 0, len($6))
    for _, col := range $6 {
      cols = append(cols, &NonStarExpr{Expr: col.Name})
      vals = append(vals, col.Expr)
    }
    $$ = &Insert{Comments: Comments($2), Table: $4, Columns: cols, Rows: Values{vals}, OnDup: OnDup($7)}
  }

update_statement:
  UPDATE comment_opt dml_table_expression SET update_list where_expression_opt order_by_opt limit_opt
  {
    $$ = &Update{Comments: Comments($2), Table: $3, Exprs: $5, Where: NewWhere(AST_WHERE, $6), OrderBy: $7, Limit: $8}
  }

delete_statement:
  DELETE comment_opt FROM dml_table_expression where_expression_opt order_by_opt limit_opt
  {
    $$ = &Delete{Comments: Comments($2), Table: $4, Where: NewWhere(AST_WHERE, $5), OrderBy: $6, Limit: $7}
  }

set_statement:
  SET comment_opt update_list
  {
    $$ = &Set{Comments: Comments($2), Exprs: $3}
  }

zero_fill_opt:
  {
    $$ = ""
  }
| ZEROFILL
  {
    $$ = AST_ZEROFILL
  }
data_type:
  numeric_type unsigned_opt zero_fill_opt
  {
    $$ = $1
    if $2 != "" {
        $$ += " " + $2
    }
    if $3 != "" {
        $$ += " " + $3
    }
  }
| char_type
| time_type

time_type:
  DATE
  {
    $$ = AST_DATE
  }
| TIME
  {
    $$ = AST_TIME
  }
| TIMESTAMP
  {
    $$ = AST_TIMESTAMP
  }
| DATETIME
  {
    $$ = AST_DATETIME
  }
| YEAR
  {
    $$ = AST_YEAR
  }

char_type:
  CHAR length_opt
  {
    if $2 == "" {
        $$ = AST_CHAR
    } else {
        $$ = AST_CHAR + $2
    }
  }
| VARCHAR length_opt
  {
    if $2 == "" {
        $$ = AST_VARCHAR
    } else {
        $$ = AST_VARCHAR + $2
    }
  }
| TEXT
  {
    $$ = AST_TEXT
  }
| MEDIUMTEXT
  {
    $$ = AST_MEDIUMTEXT
  }
| MEDIUMTEXT CHARSET ID
  {
    $$ = AST_MEDIUMTEXT
    // do something with the charset id?
  }

numeric_type:
  int_type length_opt
  {
    $$ = $1 + $2
  }
| decimal_type
  {
    $$ = $1
  }

int_type:
  BIT
  {
    $$ = AST_BIT
  }
| TINYINT
  {
    $$ = AST_TINYINT
  }
| SMALLINT
  {
    $$ = AST_SMALLINT
  }
| MEDIUMINT
  {
    $$ = AST_MEDIUMINT
  }
| INT
  {
    $$ = AST_INT
  }
| INTEGER
  {
    $$ = AST_INTEGER
  }
| BIGINT
  {
    $$ = AST_BIGINT
  }

decimal_type:
  REAL precision_opt
  {
    $$ = AST_REAL + $2
  }
| DOUBLE precision_opt
  {
    $$ = AST_DOUBLE + $2
  }
| FLOAT precision_opt
  {
    $$ = AST_FLOAT + $2
  }
| DECIMAL precision_opt
  {
    $$ = AST_DECIMAL + $2
  }
| DECIMAL length_opt
  {
    $$ = AST_DECIMAL + $2
  }
| NUMERIC precision_opt
  {
    $$ = AST_NUMERIC + $2
  }
| NUMERIC length_opt
  {
    $$ = AST_NUMERIC + $2
  }

precision_opt:
  {
    $$ = ""
  }
| '(' NUMBER ',' NUMBER ')'
  {
    $$ = "(" + string($2) + ", " + string($4) + ")"
  }

length_opt:
  {
    $$ = ""
  }
| '(' NUMBER ')'
  {
    $$ = "(" + string($2) + ")"
  }

unsigned_opt:
  {
    $$ = ""
  }
| UNSIGNED
  {
    $$ = AST_UNSIGNED
  }

column_atts:
  {
    $$ = ColumnAtts{}
  }
| column_atts NOT NULL
  {
    $$ = append($$, AST_NOT_NULL)
  }

| column_atts NULL
| column_atts DEFAULT STRING
  {
    node := StrVal($3)
    $$ = append($$, "default " + String(node))
  }
| column_atts DEFAULT NUMBER
  {
    node := NumVal($3)
    $$ = append($$, "default " + String(node))
  }
| column_atts AUTO_INCREMENT
  {
    $$ = append($$, AST_AUTO_INCREMENT)
  }
| column_atts key_att
{
    $$ = append($$, $2)
}

key_att:
  primary_key
  { 
    $$ = AST_PRIMARY_KEY
  }
| unique_key
  {
    $$ = AST_UNIQUE_KEY
  }

primary_key:
  PRIMARY KEY
| KEY 

unique_key:
  UNIQUE
| UNIQUE KEY

column_definition:
  ID data_type column_atts
  {
    $$ = &ColumnDefinition{ColName: string($1), ColType: $2, ColumnAtts: $3  }
  }
  
column_definition_list:
  column_definition
  {
    $$ = ColumnDefinitions{$1}
  }
| column_definition_list ',' column_definition
  {
    $$ = append($$, $3)
  }

create_table_statement:
  CREATE TABLE not_exists_opt ID '(' column_definition_list  ')' engine_opt
  {
    $$ = &CreateTable{Name: $4, ColumnDefinitions: $6}
  }

engine_opt:
  /* nothing */
  | ENGINE '=' ID

create_statement:
  create_table_statement
  {
    $$ = $1
  }
| CREATE constraint_opt INDEX ID using_opt ON ID force_eof
  {
    // Change this to an alter statement
    $$ = &DDL{Action: AST_ALTER, Table: $7, NewName: $7}
  }
| CREATE VIEW ID force_eof
  {
    $$ = &DDL{Action: AST_CREATE, NewName: $3}
  }

alter_statement:
  ALTER ignore_opt TABLE ID non_rename_operation force_eof
  {
    $$ = &DDL{Action: AST_ALTER, Table: $4, NewName: $4}
  }
| ALTER ignore_opt TABLE ID RENAME to_opt ID
  {
    // Change this to a rename statement
    $$ = &DDL{Action: AST_RENAME, Table: $4, NewName: $7}
  }
| ALTER VIEW ID force_eof
  {
    $$ = &DDL{Action: AST_ALTER, Table: $3, NewName: $3}
  }

rename_statement:
  RENAME TABLE ID TO ID
  {
    $$ = &DDL{Action: AST_RENAME, Table: $3, NewName: $5}
  }

drop_statement:
  DROP TABLE exists_opt ID
  {
    $$ = &DDL{Action: AST_DROP, Table: $4}
  }
| DROP INDEX ID ON ID
  {
    // Change this to an alter statement
    $$ = &DDL{Action: AST_ALTER, Table: $5, NewName: $5}
  }
| DROP VIEW exists_opt ID force_eof
  {
    $$ = &DDL{Action: AST_DROP, Table: $4}
  }

analyze_statement:
  ANALYZE TABLE ID
  {
    $$ = &DDL{Action: AST_ALTER, Table: $3, NewName: $3}
  }

comment_opt:
  {
    SetAllowComments(yylex, true)
  }
  comment_list
  {
    $$ = $2
    SetAllowComments(yylex, false)
  }

comment_list:
  {
    $$ = nil
  }
| comment_list COMMENT
  {
    $$ = append($1, $2)
  }

union_op:
  UNION
  {
    $$ = AST_UNION
  }
| UNION ALL
  {
    $$ = AST_UNION_ALL
  }
| MINUS
  {
    $$ = AST_SET_MINUS
  }
| EXCEPT
  {
    $$ = AST_EXCEPT
  }
| INTERSECT
  {
    $$ = AST_INTERSECT
  }

distinct_opt:
  {
    $$ = ""
  }
| DISTINCT
  {
    $$ = AST_DISTINCT
  }

select_expression_list:
  select_expression
  {
    $$ = SelectExprs{$1}
  }
| select_expression_list ',' select_expression
  {
    $$ = append($$, $3)
  }

select_expression:
  '*'
  {
    $$ = &StarExpr{}
  }
| expression as_lower_opt
  {
    $$ = &NonStarExpr{Expr: $1, As: $2}
  }
| ID '.' '*'
  {
    $$ = &StarExpr{TableName: $1}
  }

expression:
  boolean_expression
  {
    $$ = $1
  }
| value_expression
  {
    $$ = $1
  }

as_lower_opt:
  {
    $$ = nil
  }
| ID
  {
    $$ = $1
  }
| AS ID
  {
    $$ = $2
  }

table_expression_list:
  table_expression
  {
    $$ = TableExprs{$1}
  }
| table_expression_list ',' table_expression
  {
    $$ = append($$, $3)
  }

table_expression:
  simple_table_expression as_opt index_hint_list
  {
    $$ = &AliasedTableExpr{Expr:$1, As: $2, Hints: $3}
  }
| '(' table_expression ')'
  {
    $$ = &ParenTableExpr{Expr: $2}
  }
| table_expression join_type table_expression %prec JOIN
  {
    $$ = &JoinTableExpr{LeftExpr: $1, Join: $2, RightExpr: $3}
  }
| table_expression join_type table_expression ON boolean_expression %prec JOIN
  {
    $$ = &JoinTableExpr{LeftExpr: $1, Join: $2, RightExpr: $3, On: $5}
  }

as_opt:
  {
    $$ = nil
  }
| ID
  {
    $$ = $1
  }
| AS ID
  {
    $$ = $2
  }

join_type:
  JOIN
  {
    $$ = AST_JOIN
  }
| STRAIGHT_JOIN
  {
    $$ = AST_STRAIGHT_JOIN
  }
| LEFT JOIN
  {
    $$ = AST_LEFT_JOIN
  }
| LEFT OUTER JOIN
  {
    $$ = AST_LEFT_JOIN
  }
| RIGHT JOIN
  {
    $$ = AST_RIGHT_JOIN
  }
| RIGHT OUTER JOIN
  {
    $$ = AST_RIGHT_JOIN
  }
| INNER JOIN
  {
    $$ = AST_INNER_JOIN
  }
| CROSS JOIN
  {
    $$ = AST_CROSS_JOIN
  }
| NATURAL JOIN
  {
    $$ = AST_NATURAL_JOIN
  }

simple_table_expression:
ID
  {
    $$ = &TableName{Name: $1}
  }
| ID '.' id_or_reserved_word
  {
    $$ = &TableName{Qualifier: $1, Name: $3}
  }
| subquery
  {
    $$ = $1
  }

dml_table_expression:
ID
  {
    $$ = &TableName{Name: $1}
  }
| ID '.' id_or_reserved_word
  {
    $$ = &TableName{Qualifier: $1, Name: $3}
  }

index_hint_list:
  {
    $$ = nil
  }
| USE INDEX '(' index_list ')'
  {
    $$ = &IndexHints{Type: AST_USE, Indexes: $4}
  }
| IGNORE INDEX '(' index_list ')'
  {
    $$ = &IndexHints{Type: AST_IGNORE, Indexes: $4}
  }
| FORCE INDEX '(' index_list ')'
  {
    $$ = &IndexHints{Type: AST_FORCE, Indexes: $4}
  }

index_list:
  ID
  {
    $$ = [][]rune{$1}
  }
| index_list ',' ID
  {
    $$ = append($1, $3)
  }

where_expression_opt:
  {
    $$ = nil
  }
| WHERE boolean_expression
  {
    $$ = $2
  }

boolean_expression:
  condition
| boolean_expression AND boolean_expression
  {
    $$ = &AndExpr{Left: $1, Right: $3}
  }
| boolean_expression OR boolean_expression
  {
    $$ = &OrExpr{Left: $1, Right: $3}
  }
| NOT boolean_expression
  {
    $$ = &NotExpr{Expr: $2}
  }
| '(' boolean_expression ')'
  {
    $$ = &ParenBoolExpr{Expr: $2}
  }

condition:
  value_expression compare value_expression
  {
    $$ = &ComparisonExpr{Left: $1, Operator: $2, Right: $3}
  }
| value_expression IN col_tuple
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_IN, Right: $3}
  }
| value_expression NOT IN col_tuple
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_NOT_IN, Right: $4}
  }
| value_expression LIKE value_expression
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_LIKE, Right: $3}
  }
| value_expression NOT LIKE value_expression
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_NOT_LIKE, Right: $4}
  }
| value_expression BETWEEN value_expression AND value_expression
  {
    $$ = &RangeCond{Left: $1, Operator: AST_BETWEEN, From: $3, To: $5}
  }
| value_expression NOT BETWEEN value_expression AND value_expression
  {
    $$ = &RangeCond{Left: $1, Operator: AST_NOT_BETWEEN, From: $4, To: $6}
  }
| value_expression IS value_expression
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_IS, Right: $3}
  }
| value_expression IS NOT value_expression
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_IS_NOT, Right: $4}
  }
| EXISTS subquery
  {
    $$ = &ExistsExpr{Subquery: $2}
  }

compare:
  '='
  {
    $$ = AST_EQ
  }
| '<'
  {
    $$ = AST_LT
  }
| '>'
  {
    $$ = AST_GT
  }
| LE
  {
    $$ = AST_LE
  }
| GE
  {
    $$ = AST_GE
  }
| NE
  {
    $$ = AST_NE
  }
| NULL_SAFE_EQUAL
  {
    $$ = AST_NSE
  }

col_tuple:
  '(' value_expression_list ')'
  {
    $$ = ValTuple($2)
  }
| subquery
  {
    $$ = $1
  }
| LIST_ARG
  {
    $$ = ListArg($1)
  }

subquery:
  '(' select_statement ')'
  {
    $$ = &Subquery{$2}
  }

value_expression_list:
  value_expression
  {
    $$ = ValExprs{$1}
  }
| value_expression_list ',' value_expression
  {
    $$ = append($1, $3)
  }

value_expression:
  value
  {
    $$ = $1
  }
| column_name
  {
    $$ = $1
  }
| row_tuple
  {
    $$ = $1
  }
| value_expression '&' value_expression
  {
    $$ = &BinaryExpr{Left: $1, Operator: AST_BITAND, Right: $3}
  }
| value_expression '|' value_expression
  {
    $$ = &BinaryExpr{Left: $1, Operator: AST_BITOR, Right: $3}
  }
| value_expression '^' value_expression
  {
    $$ = &BinaryExpr{Left: $1, Operator: AST_BITXOR, Right: $3}
  }
| value_expression '+' value_expression
  {
    $$ = &BinaryExpr{Left: $1, Operator: AST_PLUS, Right: $3}
  }
| value_expression '-' value_expression
  {
    $$ = &BinaryExpr{Left: $1, Operator: AST_MINUS, Right: $3}
  }
| value_expression '*' value_expression
  {
    $$ = &BinaryExpr{Left: $1, Operator: AST_MULT, Right: $3}
  }
| value_expression '/' value_expression
  {
    $$ = &BinaryExpr{Left: $1, Operator: AST_DIV, Right: $3}
  }
| value_expression '%' value_expression
  {
    $$ = &BinaryExpr{Left: $1, Operator: AST_MOD, Right: $3}
  }
/*
| value_expression OR value_expression
  {
    $$ = &BinaryExpr{Left: $1, Operator: AST_OR, Right: $3}
  }
| value_expression AND value_expression
  {
    $$ = &BinaryExpr{Left: $1, Operator: AST_AND, Right: $3}
  }
| value_expression IS value_expression
  {
    $$ = &BinaryExpr{Left: $1, Operator: AST_IS, Right: $3}
  }
| value_expression IS NOT value_expression
  {
    $$ = &BinaryExpr{Left: $1, Operator: AST_IS_NOT, Right: $4}
  }
*/
| unary_operator value_expression %prec UNARY
  {
    if num, ok := $2.(NumVal); ok {
      switch $1 {
      case '-':
        $$ = append(NumVal("-"), num...)
      case '+':
        $$ = num
      default:
        $$ = &UnaryExpr{Operator: $1, Expr: $2}
      }
    } else {
      $$ = &UnaryExpr{Operator: $1, Expr: $2}
    }
  }
| ID '(' ')'
  {
    $$ = &FuncExpr{Name: $1}
  }
| ID '(' select_expression_list ')'
  {
    $$ = &FuncExpr{Name: $1, Exprs: $3}
  }
| ID '(' DISTINCT select_expression_list ')'
  {
    $$ = &FuncExpr{Name: $1, Distinct: true, Exprs: $4}
  }
| keyword_as_func '(' select_expression_list ')'
  {
    $$ = &FuncExpr{Name: $1, Exprs: $3}
  }
| keyword_as_func '(' select_statement ')'
  {
    // XXX(toshok) select_statement is dropped here
    $$ = &FuncExpr{Name: $1}
  }
| case_expression
  {
    $$ = $1
  }

keyword_as_func:
  IF
  {
    $$ = $1
  }
| VALUES
  {
    $$ = $1
  }

unary_operator:
  '+'
  {
    $$ = AST_UPLUS
  }
| '-'
  {
    $$ = AST_UMINUS
  }
| '~'
  {
    $$ = AST_TILDA
  }

case_expression:
  CASE value_expression_opt when_expression_list else_expression_opt END
  {
    $$ = &CaseExpr{Expr: $2, Whens: $3, Else: $4}
  }

value_expression_opt:
  {
    $$ = nil
  }
| value_expression
  {
    $$ = $1
  }

when_expression_list:
  when_expression
  {
    $$ = []*When{$1}
  }
| when_expression_list when_expression
  {
    $$ = append($1, $2)
  }

when_expression:
  WHEN boolean_expression THEN value_expression
  {
    $$ = &When{Cond: $2, Val: $4}
  }

else_expression_opt:
  {
    $$ = nil
  }
| ELSE value_expression
  {
    $$ = $2
  }

column_name:
  ID
  {
    $$ = &ColName{Name: $1}
  }
| ID '.' id_or_reserved_word
  {
    $$ = &ColName{Qualifier: $1, Name: $3}
  }

value:
  STRING
  {
    $$ = StrVal($1)
  }
| NUMBER
  {
    $$ = NumVal($1)
  }
| VALUE_ARG
  {
    $$ = ValArg($1)
  }
| NULL
  {
    $$ = &NullVal{}
  }
| BINARY STRING
  {
    $$ = BinaryVal($2)
  }
| TIMESTAMP STRING
  {
    $$ = TimestampVal($2)
  }

group_by_opt:
  {
    $$ = nil
  }
| GROUP BY value_expression_list
  {
    $$ = $3
  }

having_opt:
  {
    $$ = nil
  }
| HAVING boolean_expression
  {
    $$ = $2
  }

order_by_opt:
  {
    $$ = nil
  }
| ORDER BY order_list
  {
    $$ = $3
  }

order_list:
  order
  {
    $$ = OrderBy{$1}
  }
| order_list ',' order
  {
    $$ = append($1, $3)
  }

order:
  value_expression asc_desc_opt
  {
    $$ = &Order{Expr: $1, Direction: $2}
  }

asc_desc_opt:
  {
    $$ = AST_ASC
  }
| ASC
  {
    $$ = AST_ASC
  }
| DESC
  {
    $$ = AST_DESC
  }

limit_opt:
  {
    $$ = nil
  }
| LIMIT value_expression
  {
    $$ = &Limit{Rowcount: $2}
  }
| LIMIT value_expression ',' value_expression
  {
    $$ = &Limit{Offset: $2, Rowcount: $4}
  }
| LIMIT value_expression OFFSET value_expression
  {
    $$ = &Limit{Offset: $4, Rowcount: $2}
  }

lock_opt:
  {
    $$ = ""
  }
| FOR UPDATE
  {
    $$ = AST_FOR_UPDATE
  }
| LOCK IN ID ID
  {
    if string($3) != SHARE {
      yylex.Error("expecting share")
      return 1
    }
    if string($4) != MODE {
      yylex.Error("expecting mode")
      return 1
    }
    $$ = AST_SHARE_MODE
  }

column_list_opt:
  {
    $$ = nil
  }
| '(' column_list ')'
  {
    $$ = $2
  }

column_list:
  column_name
  {
    $$ = Columns{&NonStarExpr{Expr: $1}}
  }
| column_list ',' column_name
  {
    $$ = append($$, &NonStarExpr{Expr: $3})
  }

on_dup_opt:
  {
    $$ = nil
  }
| ON DUPLICATE KEY UPDATE update_list
  {
    $$ = $5
  }

row_list:
  VALUES tuple_list
  {
    $$ = $2
  }
| select_statement
  {
    $$ = $1
  }

tuple_list:
  row_tuple
  {
    $$ = Values{$1}
  }
| tuple_list ',' row_tuple
  {
    $$ = append($1, $3)
  }

row_tuple:
  '(' value_expression_list ')'
  {
    $$ = ValTuple($2)
  }
| subquery
  {
    $$ = $1
  }

update_list:
  update_expression
  {
    $$ = UpdateExprs{$1}
  }
| update_list ',' update_expression
  {
    $$ = append($1, $3)
  }

update_expression:
  column_name '=' value_expression
  {
    $$ = &UpdateExpr{Name: $1, Expr: $3} 
  }

exists_opt:
  { $$ = struct{}{} }
| IF EXISTS
  { $$ = struct{}{} }

not_exists_opt:
  { $$ = struct{}{} }
| IF NOT EXISTS
  { $$ = struct{}{} }

ignore_opt:
  { $$ = struct{}{} }
| IGNORE
  { $$ = struct{}{} }

non_rename_operation:
  ALTER
  { $$ = struct{}{} }
| DEFAULT
  { $$ = struct{}{} }
| DROP
  { $$ = struct{}{} }
| ORDER
  { $$ = struct{}{} }
| ID
  { $$ = struct{}{} }

to_opt:
  { $$ = struct{}{} }
| TO
  { $$ = struct{}{} }

constraint_opt:
  { $$ = struct{}{} }
| UNIQUE
  { $$ = struct{}{} }

using_opt:
  { $$ = struct{}{} }
| USING ID
  { $$ = struct{}{} }

id_or_reserved_word:
  ID { $$ = $1 }
| SELECT { $$ = $1 }
| INSERT { $$ = $1 }
| UPDATE { $$ = $1 }
| DELETE { $$ = $1 }
| FROM { $$ = $1 }
| WHERE { $$ = $1 }
| GROUP { $$ = $1 }
| HAVING { $$ = $1 }
| ORDER { $$ = $1 }
| BY { $$ = $1 }
| LIMIT { $$ = $1 }
| FOR { $$ = $1 }
| OFFSET { $$ = $1 }
| ALL { $$ = $1 }
| DISTINCT { $$ = $1 }
| AS { $$ = $1 }
| EXISTS { $$ = $1 }
| IN { $$ = $1 }
| IS { $$ = $1 }
| LIKE { $$ = $1 }
| BETWEEN { $$ = $1 }
| NULL { $$ = $1 }
| ASC { $$ = $1 }
| DESC { $$ = $1 }
| VALUES { $$ = $1 }
| INTO { $$ = $1 }
| DUPLICATE { $$ = $1 }
| KEY { $$ = $1 }
| DEFAULT { $$ = $1 }
| SET { $$ = $1 }
| LOCK { $$ = $1 }
| BINARY { $$ = $1 }
| PRIMARY { $$ = $1 }
| UNIQUE { $$ = $1 }
| UNION { $$ = $1 }
| MINUS { $$ = $1 }
| EXCEPT { $$ = $1 }
| INTERSECT { $$ = $1 }
| JOIN { $$ = $1 }
| STRAIGHT_JOIN { $$ = $1 }
| LEFT { $$ = $1 }
| RIGHT { $$ = $1 }
| INNER { $$ = $1 }
| OUTER { $$ = $1 }
| CROSS { $$ = $1 }
| NATURAL { $$ = $1 }
| USE { $$ = $1 }
| FORCE { $$ = $1 }
| ON { $$ = $1 }
| OR { $$ = $1 }
| AND { $$ = $1 }
| NOT { $$ = $1 }
| CASE { $$ = $1 }
| WHEN { $$ = $1 }
| THEN { $$ = $1 }
| ELSE { $$ = $1 }
| END { $$ = $1 }
| CREATE { $$ = $1 }
| ALTER { $$ = $1 }
| DROP { $$ = $1 }
| RENAME { $$ = $1 }
| ANALYZE { $$ = $1 }
| ENGINE { $$ = $1 }
| TABLE { $$ = $1 }
| INDEX { $$ = $1 }
| VIEW { $$ = $1 }
| TO { $$ = $1 }
| IGNORE { $$ = $1 }
| IF { $$ = $1 }
| USING { $$ = $1 }
| SHOW { $$ = $1 }
| DESCRIBE { $$ = $1 }
| EXPLAIN { $$ = $1 }

force_eof:
{
  ForceEOF(yylex)
}
