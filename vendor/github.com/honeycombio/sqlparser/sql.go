//line sql.y:6
package sqlparser

import __yyfmt__ "fmt"

//line sql.y:6
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

//line sql.y:27
type yySymType struct {
	yys         int
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
	createTableStmt   CreateTable
	columnDefinition  *ColumnDefinition
	columnDefinitions ColumnDefinitions
	columnAtts        ColumnAtts
}

const LEX_ERROR = 57346
const SELECT = 57347
const INSERT = 57348
const UPDATE = 57349
const DELETE = 57350
const FROM = 57351
const WHERE = 57352
const GROUP = 57353
const HAVING = 57354
const ORDER = 57355
const BY = 57356
const LIMIT = 57357
const FOR = 57358
const OFFSET = 57359
const ALL = 57360
const DISTINCT = 57361
const AS = 57362
const EXISTS = 57363
const IN = 57364
const IS = 57365
const LIKE = 57366
const BETWEEN = 57367
const NULL = 57368
const ASC = 57369
const DESC = 57370
const VALUES = 57371
const INTO = 57372
const DUPLICATE = 57373
const KEY = 57374
const DEFAULT = 57375
const SET = 57376
const LOCK = 57377
const BINARY = 57378
const ID = 57379
const STRING = 57380
const NUMBER = 57381
const VALUE_ARG = 57382
const LIST_ARG = 57383
const COMMENT = 57384
const LE = 57385
const GE = 57386
const NE = 57387
const NULL_SAFE_EQUAL = 57388
const PRIMARY = 57389
const UNIQUE = 57390
const UNION = 57391
const MINUS = 57392
const EXCEPT = 57393
const INTERSECT = 57394
const JOIN = 57395
const STRAIGHT_JOIN = 57396
const LEFT = 57397
const RIGHT = 57398
const INNER = 57399
const OUTER = 57400
const CROSS = 57401
const NATURAL = 57402
const USE = 57403
const FORCE = 57404
const ON = 57405
const OR = 57406
const AND = 57407
const NOT = 57408
const UNARY = 57409
const CASE = 57410
const WHEN = 57411
const THEN = 57412
const ELSE = 57413
const END = 57414
const CREATE = 57415
const ALTER = 57416
const DROP = 57417
const RENAME = 57418
const ANALYZE = 57419
const ENGINE = 57420
const TABLE = 57421
const INDEX = 57422
const VIEW = 57423
const TO = 57424
const IGNORE = 57425
const IF = 57426
const USING = 57427
const SHOW = 57428
const DESCRIBE = 57429
const EXPLAIN = 57430
const BIT = 57431
const TINYINT = 57432
const SMALLINT = 57433
const MEDIUMINT = 57434
const INT = 57435
const INTEGER = 57436
const BIGINT = 57437
const REAL = 57438
const DOUBLE = 57439
const FLOAT = 57440
const UNSIGNED = 57441
const ZEROFILL = 57442
const DECIMAL = 57443
const NUMERIC = 57444
const DATE = 57445
const TIME = 57446
const TIMESTAMP = 57447
const DATETIME = 57448
const YEAR = 57449
const TEXT = 57450
const CHAR = 57451
const VARCHAR = 57452
const MEDIUMTEXT = 57453
const CHARSET = 57454
const NULLX = 57455
const AUTO_INCREMENT = 57456
const BOOL = 57457
const APPROXNUM = 57458
const INTNUM = 57459

var yyToknames = [...]string{
	"$end",
	"error",
	"$unk",
	"LEX_ERROR",
	"SELECT",
	"INSERT",
	"UPDATE",
	"DELETE",
	"FROM",
	"WHERE",
	"GROUP",
	"HAVING",
	"ORDER",
	"BY",
	"LIMIT",
	"FOR",
	"OFFSET",
	"ALL",
	"DISTINCT",
	"AS",
	"EXISTS",
	"IN",
	"IS",
	"LIKE",
	"BETWEEN",
	"NULL",
	"ASC",
	"DESC",
	"VALUES",
	"INTO",
	"DUPLICATE",
	"KEY",
	"DEFAULT",
	"SET",
	"LOCK",
	"BINARY",
	"ID",
	"STRING",
	"NUMBER",
	"VALUE_ARG",
	"LIST_ARG",
	"COMMENT",
	"LE",
	"GE",
	"NE",
	"NULL_SAFE_EQUAL",
	"'('",
	"'='",
	"'<'",
	"'>'",
	"'~'",
	"PRIMARY",
	"UNIQUE",
	"UNION",
	"MINUS",
	"EXCEPT",
	"INTERSECT",
	"','",
	"JOIN",
	"STRAIGHT_JOIN",
	"LEFT",
	"RIGHT",
	"INNER",
	"OUTER",
	"CROSS",
	"NATURAL",
	"USE",
	"FORCE",
	"ON",
	"OR",
	"AND",
	"NOT",
	"'&'",
	"'|'",
	"'^'",
	"'+'",
	"'-'",
	"'*'",
	"'/'",
	"'%'",
	"'.'",
	"UNARY",
	"CASE",
	"WHEN",
	"THEN",
	"ELSE",
	"END",
	"CREATE",
	"ALTER",
	"DROP",
	"RENAME",
	"ANALYZE",
	"ENGINE",
	"TABLE",
	"INDEX",
	"VIEW",
	"TO",
	"IGNORE",
	"IF",
	"USING",
	"SHOW",
	"DESCRIBE",
	"EXPLAIN",
	"BIT",
	"TINYINT",
	"SMALLINT",
	"MEDIUMINT",
	"INT",
	"INTEGER",
	"BIGINT",
	"REAL",
	"DOUBLE",
	"FLOAT",
	"UNSIGNED",
	"ZEROFILL",
	"DECIMAL",
	"NUMERIC",
	"DATE",
	"TIME",
	"TIMESTAMP",
	"DATETIME",
	"YEAR",
	"TEXT",
	"CHAR",
	"VARCHAR",
	"MEDIUMTEXT",
	"CHARSET",
	"NULLX",
	"AUTO_INCREMENT",
	"BOOL",
	"APPROXNUM",
	"INTNUM",
	"')'",
}
var yyStatenames = [...]string{}

const yyEofCode = 1
const yyErrCode = 2
const yyInitialStackSize = 16

//line yacctab:1
var yyExca = [...]int{
	-1, 1,
	1, -1,
	-2, 0,
}

const yyNprod = 332
const yyPrivate = 57344

var yyTokenNames []string
var yyStates []string

const yyLast = 976

var yyAct = [...]int{

	152, 56, 83, 430, 365, 509, 156, 316, 84, 113,
	320, 250, 266, 357, 443, 307, 77, 438, 280, 246,
	72, 155, 3, 522, 519, 95, 330, 331, 332, 333,
	334, 506, 335, 336, 442, 484, 437, 58, 519, 25,
	26, 27, 28, 57, 519, 507, 47, 301, 129, 128,
	78, 385, 386, 387, 388, 389, 390, 391, 392, 393,
	394, 254, 73, 395, 396, 380, 381, 382, 383, 384,
	378, 376, 377, 379, 122, 115, 25, 26, 27, 28,
	68, 61, 157, 38, 368, 326, 158, 363, 116, 122,
	122, 40, 301, 41, 411, 413, 461, 151, 154, 521,
	415, 460, 166, 43, 44, 45, 170, 58, 459, 249,
	58, 299, 62, 520, 161, 64, 434, 171, 352, 518,
	506, 308, 418, 78, 35, 412, 37, 46, 78, 42,
	264, 279, 421, 341, 287, 288, 289, 291, 292, 293,
	294, 295, 296, 297, 298, 277, 278, 270, 275, 417,
	109, 248, 470, 471, 105, 300, 276, 282, 88, 369,
	303, 78, 362, 92, 351, 342, 100, 302, 129, 128,
	312, 58, 314, 93, 76, 89, 90, 91, 127, 128,
	474, 304, 305, 423, 81, 271, 481, 475, 98, 311,
	315, 308, 111, 355, 141, 142, 143, 107, 136, 137,
	138, 139, 140, 141, 142, 143, 480, 482, 358, 80,
	127, 456, 109, 96, 97, 74, 129, 128, 358, 322,
	101, 13, 13, 14, 15, 16, 473, 119, 136, 137,
	138, 139, 140, 141, 142, 143, 99, 139, 140, 141,
	142, 143, 92, 405, 126, 100, 458, 457, 406, 409,
	408, 17, 93, 153, 89, 90, 91, 94, 407, 107,
	472, 403, 121, 159, 318, 324, 404, 98, 247, 247,
	274, 467, 330, 331, 332, 333, 334, 78, 335, 336,
	301, 340, 303, 476, 507, 327, 346, 347, 344, 468,
	425, 349, 96, 97, 270, 53, 108, 343, 495, 101,
	260, 514, 350, 345, 269, 19, 20, 22, 21, 23,
	282, 122, 466, 494, 268, 99, 328, 107, 13, 360,
	258, 493, 354, 364, 261, 356, 361, 136, 137, 138,
	139, 140, 141, 142, 143, 159, 94, 25, 26, 27,
	28, 401, 402, 449, 168, 444, 283, 439, 103, 420,
	269, 106, 281, 416, 270, 270, 422, 169, 251, 162,
	268, 160, 102, 502, 503, 164, 426, 428, 431, 427,
	67, 92, 488, 487, 100, 485, 257, 259, 256, 432,
	163, 93, 153, 89, 90, 91, 527, 526, 525, 59,
	510, 500, 159, 486, 440, 441, 98, 419, 339, 136,
	137, 138, 139, 140, 141, 142, 143, 452, 445, 446,
	447, 450, 125, 448, 451, 338, 70, 290, 321, 414,
	462, 96, 97, 398, 463, 397, 323, 272, 101, 124,
	263, 262, 120, 348, 465, 136, 137, 138, 139, 140,
	141, 142, 143, 117, 99, 114, 112, 110, 54, 71,
	69, 66, 65, 63, 104, 516, 303, 505, 504, 464,
	424, 489, 491, 13, 52, 94, 501, 497, 498, 431,
	524, 490, 499, 492, 517, 284, 252, 285, 286, 50,
	118, 48, 366, 455, 367, 317, 454, 310, 400, 247,
	55, 523, 496, 13, 508, 30, 433, 479, 513, 58,
	511, 512, 173, 174, 175, 176, 177, 178, 179, 180,
	181, 182, 183, 184, 185, 186, 187, 188, 189, 190,
	191, 192, 193, 194, 195, 196, 197, 198, 199, 200,
	201, 202, 203, 204, 172, 478, 435, 373, 375, 374,
	477, 483, 436, 371, 372, 18, 319, 370, 253, 205,
	206, 207, 208, 209, 210, 29, 211, 212, 213, 214,
	215, 216, 217, 218, 219, 220, 221, 222, 223, 224,
	31, 32, 33, 34, 36, 273, 325, 255, 39, 60,
	225, 226, 227, 228, 229, 230, 231, 232, 233, 234,
	235, 236, 237, 238, 239, 240, 241, 242, 243, 244,
	245, 173, 174, 175, 176, 177, 178, 179, 180, 181,
	182, 183, 184, 185, 186, 187, 188, 189, 190, 191,
	192, 193, 194, 195, 196, 197, 198, 199, 200, 201,
	202, 203, 204, 172, 313, 167, 515, 469, 429, 453,
	399, 353, 165, 306, 87, 85, 86, 359, 205, 206,
	207, 208, 209, 210, 82, 211, 212, 213, 214, 215,
	216, 217, 218, 219, 220, 221, 222, 223, 224, 309,
	130, 79, 410, 267, 329, 265, 75, 337, 123, 225,
	226, 227, 228, 229, 230, 231, 232, 233, 234, 235,
	236, 237, 238, 239, 240, 241, 242, 243, 244, 245,
	13, 49, 24, 51, 12, 11, 10, 9, 8, 7,
	6, 5, 4, 2, 1, 0, 88, 0, 0, 0,
	0, 92, 0, 0, 100, 0, 0, 0, 0, 0,
	0, 93, 76, 89, 90, 91, 88, 0, 0, 0,
	0, 92, 81, 0, 100, 0, 98, 0, 0, 0,
	0, 93, 76, 89, 90, 91, 0, 0, 0, 0,
	0, 0, 81, 0, 0, 0, 98, 80, 0, 0,
	0, 96, 97, 74, 0, 0, 0, 0, 101, 0,
	0, 0, 0, 0, 0, 13, 0, 80, 0, 0,
	0, 96, 97, 74, 99, 0, 0, 0, 101, 0,
	0, 88, 0, 0, 0, 0, 92, 0, 0, 100,
	0, 0, 0, 0, 99, 94, 93, 153, 89, 90,
	91, 88, 0, 0, 0, 0, 92, 81, 0, 100,
	0, 98, 0, 0, 0, 94, 93, 153, 89, 90,
	91, 0, 0, 0, 0, 0, 0, 81, 0, 0,
	0, 98, 80, 0, 0, 0, 96, 97, 0, 0,
	0, 0, 0, 101, 0, 0, 0, 0, 0, 0,
	0, 0, 80, 0, 0, 0, 96, 97, 0, 99,
	0, 92, 0, 101, 100, 0, 0, 0, 0, 0,
	0, 93, 153, 89, 90, 91, 0, 0, 0, 99,
	94, 0, 159, 0, 0, 0, 98, 0, 0, 0,
	0, 0, 0, 131, 135, 133, 134, 0, 0, 0,
	94, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 96, 97, 0, 147, 148, 149, 150, 101, 144,
	145, 146, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 99, 0, 0, 0, 0, 0,
	0, 0, 0, 132, 136, 137, 138, 139, 140, 141,
	142, 143, 0, 0, 0, 94,
}
var yyPact = [...]int{

	217, -1000, -1000, 283, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, 30,
	-5, 35, 9, 33, 488, 463, -1000, -1000, -1000, 460,
	-1000, 434, 411, 481, 352, -18, 17, 416, -1000, 21,
	415, -1000, 414, -19, 413, -19, 412, -1000, -1000, 715,
	-1000, 320, 411, 420, 73, 411, 201, -1000, 248, 69,
	410, 120, 409, -1000, 408, -1000, -9, 406, 459, 158,
	395, -1000, 253, -1000, -1000, 392, 163, 146, 891, -1000,
	800, 780, -1000, -1000, -1000, 855, 314, -1000, 312, -1000,
	-1000, -1000, -1000, 342, 327, -1000, -1000, -1000, -1000, -1000,
	-1000, 855, -1000, 310, 352, 596, 479, 352, 855, 596,
	311, 455, -39, -1000, 287, -1000, 394, -1000, -1000, 393,
	-1000, 267, 715, -1000, -1000, 390, 497, 137, 800, 800,
	855, 305, 453, 855, 855, 345, 855, 855, 855, 855,
	855, 855, 855, 855, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, 891, 131, -22, 22, 34, 891, -1000, 216,
	695, -1000, 488, -1000, -1000, 37, 155, 458, 352, 352,
	259, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, 472, 800, -1000, 155,
	-1000, 381, -1000, 150, 389, -1000, -12, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, 258, 213, 378, 313, 52,
	-1000, -1000, -1000, -1000, -1000, 32, 715, -1000, 108, 155,
	-1000, 216, -1000, -1000, 305, 855, 855, 155, 362, 155,
	855, 161, 161, 161, 116, 116, -1000, -1000, -1000, -1000,
	-1000, 855, -1000, 155, 31, -15, 107, -1000, 800, 149,
	288, 283, 139, 29, -1000, 472, 467, 470, 146, 26,
	-1000, -53, 388, -1000, -1000, 386, -1000, 477, 267, 267,
	-1000, -1000, 202, 184, 199, 191, 190, 27, -1000, 382,
	-33, 596, -1000, 16, -11, -1000, 155, 326, 855, 155,
	155, -1000, -1000, 45, -1000, 855, 98, -1000, 429, 232,
	-1000, -1000, -1000, 352, 467, -1000, 855, 855, 381, 23,
	-1000, -78, -1000, -1000, 300, -1000, 300, 300, -1000, -93,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, 298, 298, 298, 296, 296, -1000, -1000, 474,
	469, 213, 142, -1000, 188, -1000, 187, -1000, -1000, -1000,
	-1000, 13, 6, 1, -1000, -1000, -1000, -1000, -1000, 855,
	155, -1000, 155, 855, 427, 288, -1000, -1000, 254, 231,
	-1000, 125, -1000, -1000, 212, 154, -80, -1000, -1000, 336,
	-1000, -1000, 356, -1000, 334, -1000, -1000, -1000, -1000, 333,
	-1000, -1000, -1000, 472, 800, 855, 800, -1000, -1000, 274,
	266, 251, 155, 155, 485, -1000, 855, 855, 855, -1000,
	-1000, -1000, 354, 440, -1000, 325, -1000, -1000, -1000, -1000,
	426, -1000, 425, -1000, -1000, -102, -1000, 226, -13, 467,
	146, 222, 146, 353, 353, 353, 352, 155, 155, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, 262, 439, -14,
	-1000, -20, -34, 201, -110, -1000, 484, 448, -1000, 351,
	-1000, -1000, -1000, -1000, 350, -1000, 349, -1000,
}
var yyPgo = [...]int{

	0, 714, 713, 21, 712, 711, 710, 709, 708, 707,
	706, 705, 704, 555, 703, 702, 701, 20, 62, 678,
	677, 676, 675, 12, 674, 673, 295, 672, 5, 19,
	16, 671, 670, 669, 654, 0, 18, 6, 647, 8,
	646, 25, 645, 2, 644, 643, 15, 642, 641, 640,
	639, 7, 638, 3, 637, 4, 636, 635, 634, 13,
	1, 43, 370, 579, 578, 577, 576, 574, 548, 11,
	9, 547, 10, 546, 545, 17, 544, 543, 542, 541,
	540, 539, 538, 14, 537, 536, 535, 497, 496, 495,
}
var yyR1 = [...]int{

	0, 1, 2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 3, 3, 3, 4, 4, 5, 6, 7,
	79, 79, 71, 71, 71, 84, 84, 84, 84, 84,
	76, 76, 76, 76, 76, 77, 77, 81, 81, 81,
	81, 81, 81, 81, 82, 82, 82, 82, 82, 82,
	82, 83, 83, 75, 75, 78, 78, 85, 85, 85,
	85, 85, 85, 85, 80, 80, 86, 86, 87, 87,
	72, 73, 73, 74, 88, 88, 8, 8, 8, 9,
	9, 9, 10, 11, 11, 11, 12, 89, 13, 14,
	14, 15, 15, 15, 15, 15, 16, 16, 17, 17,
	18, 18, 18, 21, 21, 19, 19, 19, 22, 22,
	23, 23, 23, 23, 20, 20, 20, 24, 24, 24,
	24, 24, 24, 24, 24, 24, 25, 25, 25, 26,
	26, 27, 27, 27, 27, 28, 28, 29, 29, 30,
	30, 30, 30, 30, 31, 31, 31, 31, 31, 31,
	31, 31, 31, 31, 32, 32, 32, 32, 32, 32,
	32, 36, 36, 36, 41, 37, 37, 35, 35, 35,
	35, 35, 35, 35, 35, 35, 35, 35, 35, 35,
	35, 35, 35, 35, 35, 40, 40, 42, 42, 42,
	44, 47, 47, 45, 45, 46, 48, 48, 43, 43,
	34, 34, 34, 34, 34, 34, 49, 49, 50, 50,
	51, 51, 52, 52, 53, 54, 54, 54, 55, 55,
	55, 55, 56, 56, 56, 57, 57, 58, 58, 59,
	59, 33, 33, 38, 38, 39, 39, 60, 60, 61,
	62, 62, 63, 63, 64, 64, 65, 65, 65, 65,
	65, 66, 66, 67, 67, 68, 68, 69, 69, 69,
	69, 69, 69, 69, 69, 69, 69, 69, 69, 69,
	69, 69, 69, 69, 69, 69, 69, 69, 69, 69,
	69, 69, 69, 69, 69, 69, 69, 69, 69, 69,
	69, 69, 69, 69, 69, 69, 69, 69, 69, 69,
	69, 69, 69, 69, 69, 69, 69, 69, 69, 69,
	69, 69, 69, 69, 69, 69, 69, 69, 69, 69,
	69, 69, 69, 69, 69, 69, 69, 69, 69, 69,
	69, 70,
}
var yyR2 = [...]int{

	0, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 4, 12, 3, 7, 7, 8, 7, 3,
	0, 1, 3, 1, 1, 1, 1, 1, 1, 1,
	2, 2, 1, 1, 3, 2, 1, 1, 1, 1,
	1, 1, 1, 1, 2, 2, 2, 2, 2, 2,
	2, 0, 5, 0, 3, 0, 1, 0, 3, 2,
	3, 3, 2, 2, 1, 1, 2, 1, 1, 2,
	3, 1, 3, 8, 0, 3, 1, 8, 4, 6,
	7, 4, 5, 4, 5, 5, 3, 0, 2, 0,
	2, 1, 2, 1, 1, 1, 0, 1, 1, 3,
	1, 2, 3, 1, 1, 0, 1, 2, 1, 3,
	3, 3, 3, 5, 0, 1, 2, 1, 1, 2,
	3, 2, 3, 2, 2, 2, 1, 3, 1, 1,
	3, 0, 5, 5, 5, 1, 3, 0, 2, 1,
	3, 3, 2, 3, 3, 3, 4, 3, 4, 5,
	6, 3, 4, 2, 1, 1, 1, 1, 1, 1,
	1, 3, 1, 1, 3, 1, 3, 1, 1, 1,
	3, 3, 3, 3, 3, 3, 3, 3, 2, 3,
	4, 5, 4, 4, 1, 1, 1, 1, 1, 1,
	5, 0, 1, 1, 2, 4, 0, 2, 1, 3,
	1, 1, 1, 1, 2, 2, 0, 3, 0, 2,
	0, 3, 1, 3, 2, 0, 1, 1, 0, 2,
	4, 4, 0, 2, 4, 0, 3, 1, 3, 0,
	5, 2, 1, 1, 3, 3, 1, 1, 3, 3,
	0, 2, 0, 3, 0, 1, 1, 1, 1, 1,
	1, 0, 1, 0, 1, 0, 2, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 0,
}
var yyChk = [...]int{

	-1000, -1, -2, -3, -4, -5, -6, -7, -8, -9,
	-10, -11, -12, 5, 6, 7, 8, 34, -74, 88,
	89, 91, 90, 92, -15, 54, 55, 56, 57, -13,
	-89, -13, -13, -13, -13, 94, -67, 96, 53, -64,
	96, 98, 94, 94, 95, 96, 94, -3, 18, -16,
	19, -14, 30, -26, 37, 9, -60, -61, -43, 37,
	-63, 99, 95, 37, 94, 37, 37, -62, 99, 37,
	-62, 37, -17, -18, 78, -21, 37, -30, -35, -31,
	72, 47, -34, -43, -39, -42, -40, -44, 21, 38,
	39, 40, 26, 36, 120, -41, 76, 77, 51, 99,
	29, 83, 42, -26, 34, 81, -26, 58, 48, 81,
	37, 72, 37, -70, 37, -70, 97, 37, 21, 69,
	37, 9, 58, -19, 37, 20, 81, 47, 71, 70,
	-32, 22, 72, 24, 25, 23, 73, 74, 75, 76,
	77, 78, 79, 80, 48, 49, 50, 43, 44, 45,
	46, -30, -35, 37, -30, -3, -37, -35, -35, 47,
	47, -41, 47, 38, 38, -47, -35, -57, 34, 47,
	-60, -69, 37, 5, 6, 7, 8, 9, 10, 11,
	12, 13, 14, 15, 16, 17, 18, 19, 20, 21,
	22, 23, 24, 25, 26, 27, 28, 29, 30, 31,
	32, 33, 34, 35, 36, 52, 53, 54, 55, 56,
	57, 59, 60, 61, 62, 63, 64, 65, 66, 67,
	68, 69, 70, 71, 72, 83, 84, 85, 86, 87,
	88, 89, 90, 91, 92, 93, 94, 95, 96, 97,
	98, 99, 100, 101, 102, 103, -29, 10, -61, -35,
	-69, 47, 21, -68, 100, -65, 91, 89, 33, 90,
	13, 37, 37, 37, -70, -22, -23, -25, 47, 37,
	-41, -18, 37, 78, 133, -17, 19, -30, -30, -35,
	-36, 47, -41, 41, 22, 24, 25, -35, -35, -35,
	72, -35, -35, -35, -35, -35, -35, -35, -35, 133,
	133, 58, 133, -35, -17, -3, -45, -46, 84, -33,
	29, -3, -60, -58, -43, -29, -51, 13, -30, -73,
	-72, 37, 69, 37, -70, -66, 97, -29, 58, -24,
	59, 60, 61, 62, 63, 65, 66, -20, 37, 20,
	-23, 81, 133, -17, -37, -36, -35, -35, 71, -35,
	-35, 133, 133, -48, -46, 86, -30, -59, 69, -38,
	-39, -59, 133, 58, -51, -55, 15, 14, 58, 133,
	-71, -77, -76, -84, -81, -82, 124, 125, 123, 126,
	118, 119, 120, 121, 122, 104, 105, 106, 107, 108,
	109, 110, 111, 112, 113, 116, 117, 37, 37, -49,
	11, -23, -23, 59, 64, 59, 64, 59, 59, 59,
	-27, 67, 98, 68, 37, 133, -69, 133, 133, 71,
	-35, 87, -35, 85, 31, 58, -43, -55, -35, -52,
	-53, -35, -72, -88, 93, -85, -78, 114, -75, 47,
	-75, -75, 127, -83, 47, -83, -83, -83, -75, 47,
	-83, -75, -70, -50, 12, 14, 69, 59, 59, 95,
	95, 95, -35, -35, 32, -39, 58, 17, 58, -54,
	27, 28, 48, 72, 26, 33, 129, -80, -86, -87,
	52, 32, 53, -79, 115, 39, 37, 39, 39, -51,
	-30, -37, -30, 47, 47, 47, 7, -35, -35, -53,
	37, 26, 38, 39, 32, 32, 133, 58, -55, -28,
	37, -28, -28, -60, 39, -56, 16, 35, 133, 58,
	133, 133, 133, 7, 22, 37, 37, 37,
}
var yyDef = [...]int{

	0, -2, 1, 2, 3, 4, 5, 6, 7, 8,
	9, 10, 11, 87, 87, 87, 87, 87, 76, 253,
	244, 0, 0, 0, 0, 91, 93, 94, 95, 96,
	89, 0, 0, 0, 0, 242, 0, 0, 254, 0,
	0, 245, 0, 240, 0, 240, 0, 14, 92, 0,
	97, 88, 0, 0, 129, 0, 19, 237, 0, 198,
	0, 0, 0, 331, 0, 331, 0, 0, 0, 0,
	0, 86, 12, 98, 100, 105, 198, 103, 104, 139,
	0, 0, 167, 168, 169, 0, 0, 184, 0, 200,
	201, 202, 203, 0, 0, 236, 187, 188, 189, 185,
	186, 191, 90, 225, 0, 0, 137, 0, 0, 0,
	0, 0, 255, 78, 0, 81, 0, 83, 241, 0,
	331, 0, 0, 101, 106, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 154, 155, 156, 157, 158, 159,
	160, 142, 0, 198, 0, 0, 0, 165, 178, 0,
	0, 153, 0, 204, 205, 0, 192, 0, 0, 0,
	137, 130, 257, 258, 259, 260, 261, 262, 263, 264,
	265, 266, 267, 268, 269, 270, 271, 272, 273, 274,
	275, 276, 277, 278, 279, 280, 281, 282, 283, 284,
	285, 286, 287, 288, 289, 290, 291, 292, 293, 294,
	295, 296, 297, 298, 299, 300, 301, 302, 303, 304,
	305, 306, 307, 308, 309, 310, 311, 312, 313, 314,
	315, 316, 317, 318, 319, 320, 321, 322, 323, 324,
	325, 326, 327, 328, 329, 330, 210, 0, 238, 239,
	199, 0, 243, 0, 0, 331, 251, 246, 247, 248,
	249, 250, 82, 84, 85, 137, 108, 114, 0, 126,
	128, 99, 107, 102, 179, 0, 0, 140, 141, 144,
	145, 0, 162, 163, 0, 0, 0, 147, 0, 151,
	0, 170, 171, 172, 173, 174, 175, 176, 177, 143,
	164, 0, 235, 165, 0, 0, 196, 193, 0, 229,
	0, 232, 229, 0, 227, 210, 218, 0, 138, 0,
	71, 0, 0, 256, 79, 0, 252, 206, 0, 0,
	117, 118, 0, 0, 0, 0, 0, 131, 115, 0,
	0, 0, 180, 0, 0, 146, 148, 0, 0, 152,
	166, 182, 183, 0, 194, 0, 0, 15, 0, 231,
	233, 16, 226, 0, 218, 18, 0, 0, 0, 74,
	57, 55, 23, 24, 53, 36, 53, 53, 32, 33,
	25, 26, 27, 28, 29, 37, 38, 39, 40, 41,
	42, 43, 51, 51, 51, 51, 51, 331, 80, 208,
	0, 109, 112, 119, 0, 121, 0, 123, 124, 125,
	110, 0, 0, 0, 116, 111, 127, 181, 161, 0,
	149, 190, 197, 0, 0, 0, 228, 17, 219, 211,
	212, 215, 72, 73, 0, 70, 20, 56, 35, 0,
	30, 31, 0, 44, 0, 45, 46, 47, 48, 0,
	49, 50, 77, 210, 0, 0, 0, 120, 122, 0,
	0, 0, 150, 195, 0, 234, 0, 0, 0, 214,
	216, 217, 0, 0, 59, 0, 62, 63, 64, 65,
	0, 67, 68, 22, 21, 0, 34, 0, 0, 218,
	209, 207, 113, 0, 0, 0, 0, 220, 221, 213,
	75, 58, 60, 61, 66, 69, 54, 0, 222, 0,
	135, 0, 0, 230, 0, 13, 0, 0, 132, 0,
	133, 134, 52, 223, 0, 136, 0, 224,
}
var yyTok1 = [...]int{

	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 80, 73, 3,
	47, 133, 78, 76, 58, 77, 81, 79, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	49, 48, 50, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 75, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 74, 3, 51,
}
var yyTok2 = [...]int{

	2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
	12, 13, 14, 15, 16, 17, 18, 19, 20, 21,
	22, 23, 24, 25, 26, 27, 28, 29, 30, 31,
	32, 33, 34, 35, 36, 37, 38, 39, 40, 41,
	42, 43, 44, 45, 46, 52, 53, 54, 55, 56,
	57, 59, 60, 61, 62, 63, 64, 65, 66, 67,
	68, 69, 70, 71, 72, 82, 83, 84, 85, 86,
	87, 88, 89, 90, 91, 92, 93, 94, 95, 96,
	97, 98, 99, 100, 101, 102, 103, 104, 105, 106,
	107, 108, 109, 110, 111, 112, 113, 114, 115, 116,
	117, 118, 119, 120, 121, 122, 123, 124, 125, 126,
	127, 128, 129, 130, 131, 132,
}
var yyTok3 = [...]int{
	0,
}

var yyErrorMessages = [...]struct {
	state int
	token int
	msg   string
}{}

//line yaccpar:1

/*	parser for yacc output	*/

var (
	yyDebug        = 0
	yyErrorVerbose = false
)

type yyLexer interface {
	Lex(lval *yySymType) int
	Error(s string)
}

type yyParser interface {
	Parse(yyLexer) int
	Lookahead() int
}

type yyParserImpl struct {
	lval  yySymType
	stack [yyInitialStackSize]yySymType
	char  int
}

func (p *yyParserImpl) Lookahead() int {
	return p.char
}

func yyNewParser() yyParser {
	return &yyParserImpl{}
}

const yyFlag = -1000

func yyTokname(c int) string {
	if c >= 1 && c-1 < len(yyToknames) {
		if yyToknames[c-1] != "" {
			return yyToknames[c-1]
		}
	}
	return __yyfmt__.Sprintf("tok-%v", c)
}

func yyStatname(s int) string {
	if s >= 0 && s < len(yyStatenames) {
		if yyStatenames[s] != "" {
			return yyStatenames[s]
		}
	}
	return __yyfmt__.Sprintf("state-%v", s)
}

func yyErrorMessage(state, lookAhead int) string {
	const TOKSTART = 4

	if !yyErrorVerbose {
		return "syntax error"
	}

	for _, e := range yyErrorMessages {
		if e.state == state && e.token == lookAhead {
			return "syntax error: " + e.msg
		}
	}

	res := "syntax error: unexpected " + yyTokname(lookAhead)

	// To match Bison, suggest at most four expected tokens.
	expected := make([]int, 0, 4)

	// Look for shiftable tokens.
	base := yyPact[state]
	for tok := TOKSTART; tok-1 < len(yyToknames); tok++ {
		if n := base + tok; n >= 0 && n < yyLast && yyChk[yyAct[n]] == tok {
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}
	}

	if yyDef[state] == -2 {
		i := 0
		for yyExca[i] != -1 || yyExca[i+1] != state {
			i += 2
		}

		// Look for tokens that we accept or reduce.
		for i += 2; yyExca[i] >= 0; i += 2 {
			tok := yyExca[i]
			if tok < TOKSTART || yyExca[i+1] == 0 {
				continue
			}
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}

		// If the default action is to accept or reduce, give up.
		if yyExca[i+1] != 0 {
			return res
		}
	}

	for i, tok := range expected {
		if i == 0 {
			res += ", expecting "
		} else {
			res += " or "
		}
		res += yyTokname(tok)
	}
	return res
}

func yylex1(lex yyLexer, lval *yySymType) (char, token int) {
	token = 0
	char = lex.Lex(lval)
	if char <= 0 {
		token = yyTok1[0]
		goto out
	}
	if char < len(yyTok1) {
		token = yyTok1[char]
		goto out
	}
	if char >= yyPrivate {
		if char < yyPrivate+len(yyTok2) {
			token = yyTok2[char-yyPrivate]
			goto out
		}
	}
	for i := 0; i < len(yyTok3); i += 2 {
		token = yyTok3[i+0]
		if token == char {
			token = yyTok3[i+1]
			goto out
		}
	}

out:
	if token == 0 {
		token = yyTok2[1] /* unknown char */
	}
	if yyDebug >= 3 {
		__yyfmt__.Printf("lex %s(%d)\n", yyTokname(token), uint(char))
	}
	return char, token
}

func yyParse(yylex yyLexer) int {
	return yyNewParser().Parse(yylex)
}

func (yyrcvr *yyParserImpl) Parse(yylex yyLexer) int {
	var yyn int
	var yyVAL yySymType
	var yyDollar []yySymType
	_ = yyDollar // silence set and not used
	yyS := yyrcvr.stack[:]

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	yystate := 0
	yyrcvr.char = -1
	yytoken := -1 // yyrcvr.char translated into internal numbering
	defer func() {
		// Make sure we report no lookahead when not parsing.
		yystate = -1
		yyrcvr.char = -1
		yytoken = -1
	}()
	yyp := -1
	goto yystack

ret0:
	return 0

ret1:
	return 1

yystack:
	/* put a state and value onto the stack */
	if yyDebug >= 4 {
		__yyfmt__.Printf("char %v in %v\n", yyTokname(yytoken), yyStatname(yystate))
	}

	yyp++
	if yyp >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyS[yyp] = yyVAL
	yyS[yyp].yys = yystate

yynewstate:
	yyn = yyPact[yystate]
	if yyn <= yyFlag {
		goto yydefault /* simple state */
	}
	if yyrcvr.char < 0 {
		yyrcvr.char, yytoken = yylex1(yylex, &yyrcvr.lval)
	}
	yyn += yytoken
	if yyn < 0 || yyn >= yyLast {
		goto yydefault
	}
	yyn = yyAct[yyn]
	if yyChk[yyn] == yytoken { /* valid shift */
		yyrcvr.char = -1
		yytoken = -1
		yyVAL = yyrcvr.lval
		yystate = yyn
		if Errflag > 0 {
			Errflag--
		}
		goto yystack
	}

yydefault:
	/* default state action */
	yyn = yyDef[yystate]
	if yyn == -2 {
		if yyrcvr.char < 0 {
			yyrcvr.char, yytoken = yylex1(yylex, &yyrcvr.lval)
		}

		/* look through exception table */
		xi := 0
		for {
			if yyExca[xi+0] == -1 && yyExca[xi+1] == yystate {
				break
			}
			xi += 2
		}
		for xi += 2; ; xi += 2 {
			yyn = yyExca[xi+0]
			if yyn < 0 || yyn == yytoken {
				break
			}
		}
		yyn = yyExca[xi+1]
		if yyn < 0 {
			goto ret0
		}
	}
	if yyn == 0 {
		/* error ... attempt to resume parsing */
		switch Errflag {
		case 0: /* brand new error */
			yylex.Error(yyErrorMessage(yystate, yytoken))
			Nerrs++
			if yyDebug >= 1 {
				__yyfmt__.Printf("%s", yyStatname(yystate))
				__yyfmt__.Printf(" saw %s\n", yyTokname(yytoken))
			}
			fallthrough

		case 1, 2: /* incompletely recovered error ... try again */
			Errflag = 3

			/* find a state where "error" is a legal shift action */
			for yyp >= 0 {
				yyn = yyPact[yyS[yyp].yys] + yyErrCode
				if yyn >= 0 && yyn < yyLast {
					yystate = yyAct[yyn] /* simulate a shift of "error" */
					if yyChk[yystate] == yyErrCode {
						goto yystack
					}
				}

				/* the current p has no shift on "error", pop stack */
				if yyDebug >= 2 {
					__yyfmt__.Printf("error recovery pops state %d\n", yyS[yyp].yys)
				}
				yyp--
			}
			/* there is no state on the stack with an error shift ... abort */
			goto ret1

		case 3: /* no shift yet; clobber input char */
			if yyDebug >= 2 {
				__yyfmt__.Printf("error recovery discards %s\n", yyTokname(yytoken))
			}
			if yytoken == yyEofCode {
				goto ret1
			}
			yyrcvr.char = -1
			yytoken = -1
			goto yynewstate /* try again in the same state */
		}
	}

	/* reduction by production yyn */
	if yyDebug >= 2 {
		__yyfmt__.Printf("reduce %v in:\n\t%v\n", yyn, yyStatname(yystate))
	}

	yynt := yyn
	yypt := yyp
	_ = yypt // guard against "declared and not used"

	yyp -= yyR2[yyn]
	// yyp is now the index of $0. Perform the default action. Iff the
	// reduced production is ε, $1 is possibly out of range.
	if yyp+1 >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyVAL = yyS[yyp+1]

	/* consult goto table to find next state */
	yyn = yyR1[yyn]
	yyg := yyPgo[yyn]
	yyj := yyg + yyS[yyp].yys + 1

	if yyj >= yyLast {
		yystate = yyAct[yyg]
	} else {
		yystate = yyAct[yyj]
		if yyChk[yystate] != -yyn {
			yystate = yyAct[yyg]
		}
	}
	// dummy call; replaced with literal code
	switch yynt {

	case 1:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:177
		{
			SetParseTree(yylex, yyDollar[1].statement)
		}
	case 2:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:183
		{
			yyVAL.statement = yyDollar[1].selStmt
		}
	case 12:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:198
		{
			yyVAL.selStmt = &Select{Comments: Comments(yyDollar[2].runes2), Distinct: yyDollar[3].str, SelectExprs: yyDollar[4].selectExprs}
		}
	case 13:
		yyDollar = yyS[yypt-12 : yypt+1]
		//line sql.y:202
		{
			yyVAL.selStmt = &Select{Comments: Comments(yyDollar[2].runes2), Distinct: yyDollar[3].str, SelectExprs: yyDollar[4].selectExprs, From: yyDollar[6].tableExprs, Where: NewWhere(AST_WHERE, yyDollar[7].boolExpr), GroupBy: GroupBy(yyDollar[8].valExprs), Having: NewWhere(AST_HAVING, yyDollar[9].boolExpr), OrderBy: yyDollar[10].orderBy, Limit: yyDollar[11].limit, Lock: yyDollar[12].str}
		}
	case 14:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:206
		{
			yyVAL.selStmt = &Union{Type: yyDollar[2].str, Left: yyDollar[1].selStmt, Right: yyDollar[3].selStmt}
		}
	case 15:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:212
		{
			yyVAL.statement = &Insert{Comments: Comments(yyDollar[2].runes2), Table: yyDollar[4].tableName, Columns: yyDollar[5].columns, Rows: yyDollar[6].insRows, OnDup: OnDup(yyDollar[7].updateExprs)}
		}
	case 16:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:216
		{
			cols := make(Columns, 0, len(yyDollar[6].updateExprs))
			vals := make(ValTuple, 0, len(yyDollar[6].updateExprs))
			for _, col := range yyDollar[6].updateExprs {
				cols = append(cols, &NonStarExpr{Expr: col.Name})
				vals = append(vals, col.Expr)
			}
			yyVAL.statement = &Insert{Comments: Comments(yyDollar[2].runes2), Table: yyDollar[4].tableName, Columns: cols, Rows: Values{vals}, OnDup: OnDup(yyDollar[7].updateExprs)}
		}
	case 17:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:228
		{
			yyVAL.statement = &Update{Comments: Comments(yyDollar[2].runes2), Table: yyDollar[3].tableName, Exprs: yyDollar[5].updateExprs, Where: NewWhere(AST_WHERE, yyDollar[6].boolExpr), OrderBy: yyDollar[7].orderBy, Limit: yyDollar[8].limit}
		}
	case 18:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:234
		{
			yyVAL.statement = &Delete{Comments: Comments(yyDollar[2].runes2), Table: yyDollar[4].tableName, Where: NewWhere(AST_WHERE, yyDollar[5].boolExpr), OrderBy: yyDollar[6].orderBy, Limit: yyDollar[7].limit}
		}
	case 19:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:240
		{
			yyVAL.statement = &Set{Comments: Comments(yyDollar[2].runes2), Exprs: yyDollar[3].updateExprs}
		}
	case 20:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:245
		{
			yyVAL.str = ""
		}
	case 21:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:249
		{
			yyVAL.str = AST_ZEROFILL
		}
	case 22:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:254
		{
			yyVAL.str = yyDollar[1].str
			if yyDollar[2].str != "" {
				yyVAL.str += " " + yyDollar[2].str
			}
			if yyDollar[3].str != "" {
				yyVAL.str += " " + yyDollar[3].str
			}
		}
	case 25:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:268
		{
			yyVAL.str = AST_DATE
		}
	case 26:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:272
		{
			yyVAL.str = AST_TIME
		}
	case 27:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:276
		{
			yyVAL.str = AST_TIMESTAMP
		}
	case 28:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:280
		{
			yyVAL.str = AST_DATETIME
		}
	case 29:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:284
		{
			yyVAL.str = AST_YEAR
		}
	case 30:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:290
		{
			if yyDollar[2].str == "" {
				yyVAL.str = AST_CHAR
			} else {
				yyVAL.str = AST_CHAR + yyDollar[2].str
			}
		}
	case 31:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:298
		{
			if yyDollar[2].str == "" {
				yyVAL.str = AST_VARCHAR
			} else {
				yyVAL.str = AST_VARCHAR + yyDollar[2].str
			}
		}
	case 32:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:306
		{
			yyVAL.str = AST_TEXT
		}
	case 33:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:310
		{
			yyVAL.str = AST_MEDIUMTEXT
		}
	case 34:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:314
		{
			yyVAL.str = AST_MEDIUMTEXT
			// do something with the charset id?
		}
	case 35:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:321
		{
			yyVAL.str = yyDollar[1].str + yyDollar[2].str
		}
	case 36:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:325
		{
			yyVAL.str = yyDollar[1].str
		}
	case 37:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:331
		{
			yyVAL.str = AST_BIT
		}
	case 38:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:335
		{
			yyVAL.str = AST_TINYINT
		}
	case 39:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:339
		{
			yyVAL.str = AST_SMALLINT
		}
	case 40:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:343
		{
			yyVAL.str = AST_MEDIUMINT
		}
	case 41:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:347
		{
			yyVAL.str = AST_INT
		}
	case 42:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:351
		{
			yyVAL.str = AST_INTEGER
		}
	case 43:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:355
		{
			yyVAL.str = AST_BIGINT
		}
	case 44:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:361
		{
			yyVAL.str = AST_REAL + yyDollar[2].str
		}
	case 45:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:365
		{
			yyVAL.str = AST_DOUBLE + yyDollar[2].str
		}
	case 46:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:369
		{
			yyVAL.str = AST_FLOAT + yyDollar[2].str
		}
	case 47:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:373
		{
			yyVAL.str = AST_DECIMAL + yyDollar[2].str
		}
	case 48:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:377
		{
			yyVAL.str = AST_DECIMAL + yyDollar[2].str
		}
	case 49:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:381
		{
			yyVAL.str = AST_NUMERIC + yyDollar[2].str
		}
	case 50:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:385
		{
			yyVAL.str = AST_NUMERIC + yyDollar[2].str
		}
	case 51:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:390
		{
			yyVAL.str = ""
		}
	case 52:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:394
		{
			yyVAL.str = "(" + string(yyDollar[2].runes) + ", " + string(yyDollar[4].runes) + ")"
		}
	case 53:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:399
		{
			yyVAL.str = ""
		}
	case 54:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:403
		{
			yyVAL.str = "(" + string(yyDollar[2].runes) + ")"
		}
	case 55:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:408
		{
			yyVAL.str = ""
		}
	case 56:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:412
		{
			yyVAL.str = AST_UNSIGNED
		}
	case 57:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:417
		{
			yyVAL.columnAtts = ColumnAtts{}
		}
	case 58:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:421
		{
			yyVAL.columnAtts = append(yyVAL.columnAtts, AST_NOT_NULL)
		}
	case 60:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:427
		{
			node := StrVal(yyDollar[3].runes)
			yyVAL.columnAtts = append(yyVAL.columnAtts, "default "+String(node))
		}
	case 61:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:432
		{
			node := NumVal(yyDollar[3].runes)
			yyVAL.columnAtts = append(yyVAL.columnAtts, "default "+String(node))
		}
	case 62:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:437
		{
			yyVAL.columnAtts = append(yyVAL.columnAtts, AST_AUTO_INCREMENT)
		}
	case 63:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:441
		{
			yyVAL.columnAtts = append(yyVAL.columnAtts, yyDollar[2].str)
		}
	case 64:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:447
		{
			yyVAL.str = AST_PRIMARY_KEY
		}
	case 65:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:451
		{
			yyVAL.str = AST_UNIQUE_KEY
		}
	case 70:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:465
		{
			yyVAL.columnDefinition = &ColumnDefinition{ColName: string(yyDollar[1].runes), ColType: yyDollar[2].str, ColumnAtts: yyDollar[3].columnAtts}
		}
	case 71:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:471
		{
			yyVAL.columnDefinitions = ColumnDefinitions{yyDollar[1].columnDefinition}
		}
	case 72:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:475
		{
			yyVAL.columnDefinitions = append(yyVAL.columnDefinitions, yyDollar[3].columnDefinition)
		}
	case 73:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:481
		{
			yyVAL.statement = &CreateTable{Name: yyDollar[4].runes, ColumnDefinitions: yyDollar[6].columnDefinitions}
		}
	case 76:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:491
		{
			yyVAL.statement = yyDollar[1].statement
		}
	case 77:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:495
		{
			// Change this to an alter statement
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[7].runes, NewName: yyDollar[7].runes}
		}
	case 78:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:500
		{
			yyVAL.statement = &DDL{Action: AST_CREATE, NewName: yyDollar[3].runes}
		}
	case 79:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:506
		{
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[4].runes, NewName: yyDollar[4].runes}
		}
	case 80:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:510
		{
			// Change this to a rename statement
			yyVAL.statement = &DDL{Action: AST_RENAME, Table: yyDollar[4].runes, NewName: yyDollar[7].runes}
		}
	case 81:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:515
		{
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[3].runes, NewName: yyDollar[3].runes}
		}
	case 82:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:521
		{
			yyVAL.statement = &DDL{Action: AST_RENAME, Table: yyDollar[3].runes, NewName: yyDollar[5].runes}
		}
	case 83:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:527
		{
			yyVAL.statement = &DDL{Action: AST_DROP, Table: yyDollar[4].runes}
		}
	case 84:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:531
		{
			// Change this to an alter statement
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[5].runes, NewName: yyDollar[5].runes}
		}
	case 85:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:536
		{
			yyVAL.statement = &DDL{Action: AST_DROP, Table: yyDollar[4].runes}
		}
	case 86:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:542
		{
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[3].runes, NewName: yyDollar[3].runes}
		}
	case 87:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:547
		{
			SetAllowComments(yylex, true)
		}
	case 88:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:551
		{
			yyVAL.runes2 = yyDollar[2].runes2
			SetAllowComments(yylex, false)
		}
	case 89:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:557
		{
			yyVAL.runes2 = nil
		}
	case 90:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:561
		{
			yyVAL.runes2 = append(yyDollar[1].runes2, yyDollar[2].runes)
		}
	case 91:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:567
		{
			yyVAL.str = AST_UNION
		}
	case 92:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:571
		{
			yyVAL.str = AST_UNION_ALL
		}
	case 93:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:575
		{
			yyVAL.str = AST_SET_MINUS
		}
	case 94:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:579
		{
			yyVAL.str = AST_EXCEPT
		}
	case 95:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:583
		{
			yyVAL.str = AST_INTERSECT
		}
	case 96:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:588
		{
			yyVAL.str = ""
		}
	case 97:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:592
		{
			yyVAL.str = AST_DISTINCT
		}
	case 98:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:598
		{
			yyVAL.selectExprs = SelectExprs{yyDollar[1].selectExpr}
		}
	case 99:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:602
		{
			yyVAL.selectExprs = append(yyVAL.selectExprs, yyDollar[3].selectExpr)
		}
	case 100:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:608
		{
			yyVAL.selectExpr = &StarExpr{}
		}
	case 101:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:612
		{
			yyVAL.selectExpr = &NonStarExpr{Expr: yyDollar[1].expr, As: yyDollar[2].runes}
		}
	case 102:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:616
		{
			yyVAL.selectExpr = &StarExpr{TableName: yyDollar[1].runes}
		}
	case 103:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:622
		{
			yyVAL.expr = yyDollar[1].boolExpr
		}
	case 104:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:626
		{
			yyVAL.expr = yyDollar[1].valExpr
		}
	case 105:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:631
		{
			yyVAL.runes = nil
		}
	case 106:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:635
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 107:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:639
		{
			yyVAL.runes = yyDollar[2].runes
		}
	case 108:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:645
		{
			yyVAL.tableExprs = TableExprs{yyDollar[1].tableExpr}
		}
	case 109:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:649
		{
			yyVAL.tableExprs = append(yyVAL.tableExprs, yyDollar[3].tableExpr)
		}
	case 110:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:655
		{
			yyVAL.tableExpr = &AliasedTableExpr{Expr: yyDollar[1].smTableExpr, As: yyDollar[2].runes, Hints: yyDollar[3].indexHints}
		}
	case 111:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:659
		{
			yyVAL.tableExpr = &ParenTableExpr{Expr: yyDollar[2].tableExpr}
		}
	case 112:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:663
		{
			yyVAL.tableExpr = &JoinTableExpr{LeftExpr: yyDollar[1].tableExpr, Join: yyDollar[2].str, RightExpr: yyDollar[3].tableExpr}
		}
	case 113:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:667
		{
			yyVAL.tableExpr = &JoinTableExpr{LeftExpr: yyDollar[1].tableExpr, Join: yyDollar[2].str, RightExpr: yyDollar[3].tableExpr, On: yyDollar[5].boolExpr}
		}
	case 114:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:672
		{
			yyVAL.runes = nil
		}
	case 115:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:676
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 116:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:680
		{
			yyVAL.runes = yyDollar[2].runes
		}
	case 117:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:686
		{
			yyVAL.str = AST_JOIN
		}
	case 118:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:690
		{
			yyVAL.str = AST_STRAIGHT_JOIN
		}
	case 119:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:694
		{
			yyVAL.str = AST_LEFT_JOIN
		}
	case 120:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:698
		{
			yyVAL.str = AST_LEFT_JOIN
		}
	case 121:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:702
		{
			yyVAL.str = AST_RIGHT_JOIN
		}
	case 122:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:706
		{
			yyVAL.str = AST_RIGHT_JOIN
		}
	case 123:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:710
		{
			yyVAL.str = AST_INNER_JOIN
		}
	case 124:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:714
		{
			yyVAL.str = AST_CROSS_JOIN
		}
	case 125:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:718
		{
			yyVAL.str = AST_NATURAL_JOIN
		}
	case 126:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:724
		{
			yyVAL.smTableExpr = &TableName{Name: yyDollar[1].runes}
		}
	case 127:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:728
		{
			yyVAL.smTableExpr = &TableName{Qualifier: yyDollar[1].runes, Name: yyDollar[3].runes}
		}
	case 128:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:732
		{
			yyVAL.smTableExpr = yyDollar[1].subquery
		}
	case 129:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:738
		{
			yyVAL.tableName = &TableName{Name: yyDollar[1].runes}
		}
	case 130:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:742
		{
			yyVAL.tableName = &TableName{Qualifier: yyDollar[1].runes, Name: yyDollar[3].runes}
		}
	case 131:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:747
		{
			yyVAL.indexHints = nil
		}
	case 132:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:751
		{
			yyVAL.indexHints = &IndexHints{Type: AST_USE, Indexes: yyDollar[4].runes2}
		}
	case 133:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:755
		{
			yyVAL.indexHints = &IndexHints{Type: AST_IGNORE, Indexes: yyDollar[4].runes2}
		}
	case 134:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:759
		{
			yyVAL.indexHints = &IndexHints{Type: AST_FORCE, Indexes: yyDollar[4].runes2}
		}
	case 135:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:765
		{
			yyVAL.runes2 = [][]rune{yyDollar[1].runes}
		}
	case 136:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:769
		{
			yyVAL.runes2 = append(yyDollar[1].runes2, yyDollar[3].runes)
		}
	case 137:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:774
		{
			yyVAL.boolExpr = nil
		}
	case 138:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:778
		{
			yyVAL.boolExpr = yyDollar[2].boolExpr
		}
	case 140:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:785
		{
			yyVAL.boolExpr = &AndExpr{Left: yyDollar[1].boolExpr, Right: yyDollar[3].boolExpr}
		}
	case 141:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:789
		{
			yyVAL.boolExpr = &OrExpr{Left: yyDollar[1].boolExpr, Right: yyDollar[3].boolExpr}
		}
	case 142:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:793
		{
			yyVAL.boolExpr = &NotExpr{Expr: yyDollar[2].boolExpr}
		}
	case 143:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:797
		{
			yyVAL.boolExpr = &ParenBoolExpr{Expr: yyDollar[2].boolExpr}
		}
	case 144:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:803
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: yyDollar[2].str, Right: yyDollar[3].valExpr}
		}
	case 145:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:807
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: AST_IN, Right: yyDollar[3].colTuple}
		}
	case 146:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:811
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: AST_NOT_IN, Right: yyDollar[4].colTuple}
		}
	case 147:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:815
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: AST_LIKE, Right: yyDollar[3].valExpr}
		}
	case 148:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:819
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: AST_NOT_LIKE, Right: yyDollar[4].valExpr}
		}
	case 149:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:823
		{
			yyVAL.boolExpr = &RangeCond{Left: yyDollar[1].valExpr, Operator: AST_BETWEEN, From: yyDollar[3].valExpr, To: yyDollar[5].valExpr}
		}
	case 150:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:827
		{
			yyVAL.boolExpr = &RangeCond{Left: yyDollar[1].valExpr, Operator: AST_NOT_BETWEEN, From: yyDollar[4].valExpr, To: yyDollar[6].valExpr}
		}
	case 151:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:831
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: AST_IS, Right: yyDollar[3].valExpr}
		}
	case 152:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:835
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: AST_IS_NOT, Right: yyDollar[4].valExpr}
		}
	case 153:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:839
		{
			yyVAL.boolExpr = &ExistsExpr{Subquery: yyDollar[2].subquery}
		}
	case 154:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:845
		{
			yyVAL.str = AST_EQ
		}
	case 155:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:849
		{
			yyVAL.str = AST_LT
		}
	case 156:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:853
		{
			yyVAL.str = AST_GT
		}
	case 157:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:857
		{
			yyVAL.str = AST_LE
		}
	case 158:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:861
		{
			yyVAL.str = AST_GE
		}
	case 159:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:865
		{
			yyVAL.str = AST_NE
		}
	case 160:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:869
		{
			yyVAL.str = AST_NSE
		}
	case 161:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:875
		{
			yyVAL.colTuple = ValTuple(yyDollar[2].valExprs)
		}
	case 162:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:879
		{
			yyVAL.colTuple = yyDollar[1].subquery
		}
	case 163:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:883
		{
			yyVAL.colTuple = ListArg(yyDollar[1].runes)
		}
	case 164:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:889
		{
			yyVAL.subquery = &Subquery{yyDollar[2].selStmt}
		}
	case 165:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:895
		{
			yyVAL.valExprs = ValExprs{yyDollar[1].valExpr}
		}
	case 166:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:899
		{
			yyVAL.valExprs = append(yyDollar[1].valExprs, yyDollar[3].valExpr)
		}
	case 167:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:905
		{
			yyVAL.valExpr = yyDollar[1].valExpr
		}
	case 168:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:909
		{
			yyVAL.valExpr = yyDollar[1].colName
		}
	case 169:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:913
		{
			yyVAL.valExpr = yyDollar[1].rowTuple
		}
	case 170:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:917
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_BITAND, Right: yyDollar[3].valExpr}
		}
	case 171:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:921
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_BITOR, Right: yyDollar[3].valExpr}
		}
	case 172:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:925
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_BITXOR, Right: yyDollar[3].valExpr}
		}
	case 173:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:929
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_PLUS, Right: yyDollar[3].valExpr}
		}
	case 174:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:933
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_MINUS, Right: yyDollar[3].valExpr}
		}
	case 175:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:937
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_MULT, Right: yyDollar[3].valExpr}
		}
	case 176:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:941
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_DIV, Right: yyDollar[3].valExpr}
		}
	case 177:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:945
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_MOD, Right: yyDollar[3].valExpr}
		}
	case 178:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:967
		{
			if num, ok := yyDollar[2].valExpr.(NumVal); ok {
				switch yyDollar[1].run {
				case '-':
					yyVAL.valExpr = append(NumVal("-"), num...)
				case '+':
					yyVAL.valExpr = num
				default:
					yyVAL.valExpr = &UnaryExpr{Operator: yyDollar[1].run, Expr: yyDollar[2].valExpr}
				}
			} else {
				yyVAL.valExpr = &UnaryExpr{Operator: yyDollar[1].run, Expr: yyDollar[2].valExpr}
			}
		}
	case 179:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:982
		{
			yyVAL.valExpr = &FuncExpr{Name: yyDollar[1].runes}
		}
	case 180:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:986
		{
			yyVAL.valExpr = &FuncExpr{Name: yyDollar[1].runes, Exprs: yyDollar[3].selectExprs}
		}
	case 181:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:990
		{
			yyVAL.valExpr = &FuncExpr{Name: yyDollar[1].runes, Distinct: true, Exprs: yyDollar[4].selectExprs}
		}
	case 182:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:994
		{
			yyVAL.valExpr = &FuncExpr{Name: yyDollar[1].runes, Exprs: yyDollar[3].selectExprs}
		}
	case 183:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:998
		{
			// XXX(toshok) select_statement is dropped here
			yyVAL.valExpr = &FuncExpr{Name: yyDollar[1].runes}
		}
	case 184:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1003
		{
			yyVAL.valExpr = yyDollar[1].caseExpr
		}
	case 185:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1009
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 186:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1013
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 187:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1019
		{
			yyVAL.run = AST_UPLUS
		}
	case 188:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1023
		{
			yyVAL.run = AST_UMINUS
		}
	case 189:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1027
		{
			yyVAL.run = AST_TILDA
		}
	case 190:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1033
		{
			yyVAL.caseExpr = &CaseExpr{Expr: yyDollar[2].valExpr, Whens: yyDollar[3].whens, Else: yyDollar[4].valExpr}
		}
	case 191:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1038
		{
			yyVAL.valExpr = nil
		}
	case 192:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1042
		{
			yyVAL.valExpr = yyDollar[1].valExpr
		}
	case 193:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1048
		{
			yyVAL.whens = []*When{yyDollar[1].when}
		}
	case 194:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1052
		{
			yyVAL.whens = append(yyDollar[1].whens, yyDollar[2].when)
		}
	case 195:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1058
		{
			yyVAL.when = &When{Cond: yyDollar[2].boolExpr, Val: yyDollar[4].valExpr}
		}
	case 196:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1063
		{
			yyVAL.valExpr = nil
		}
	case 197:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1067
		{
			yyVAL.valExpr = yyDollar[2].valExpr
		}
	case 198:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1073
		{
			yyVAL.colName = &ColName{Name: yyDollar[1].runes}
		}
	case 199:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1077
		{
			yyVAL.colName = &ColName{Qualifier: yyDollar[1].runes, Name: yyDollar[3].runes}
		}
	case 200:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1083
		{
			yyVAL.valExpr = StrVal(yyDollar[1].runes)
		}
	case 201:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1087
		{
			yyVAL.valExpr = NumVal(yyDollar[1].runes)
		}
	case 202:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1091
		{
			yyVAL.valExpr = ValArg(yyDollar[1].runes)
		}
	case 203:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1095
		{
			yyVAL.valExpr = &NullVal{}
		}
	case 204:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1099
		{
			yyVAL.valExpr = BinaryVal(yyDollar[2].runes)
		}
	case 205:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1103
		{
			yyVAL.valExpr = TimestampVal(yyDollar[2].runes)
		}
	case 206:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1108
		{
			yyVAL.valExprs = nil
		}
	case 207:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1112
		{
			yyVAL.valExprs = yyDollar[3].valExprs
		}
	case 208:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1117
		{
			yyVAL.boolExpr = nil
		}
	case 209:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1121
		{
			yyVAL.boolExpr = yyDollar[2].boolExpr
		}
	case 210:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1126
		{
			yyVAL.orderBy = nil
		}
	case 211:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1130
		{
			yyVAL.orderBy = yyDollar[3].orderBy
		}
	case 212:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1136
		{
			yyVAL.orderBy = OrderBy{yyDollar[1].order}
		}
	case 213:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1140
		{
			yyVAL.orderBy = append(yyDollar[1].orderBy, yyDollar[3].order)
		}
	case 214:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1146
		{
			yyVAL.order = &Order{Expr: yyDollar[1].valExpr, Direction: yyDollar[2].str}
		}
	case 215:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1151
		{
			yyVAL.str = AST_ASC
		}
	case 216:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1155
		{
			yyVAL.str = AST_ASC
		}
	case 217:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1159
		{
			yyVAL.str = AST_DESC
		}
	case 218:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1164
		{
			yyVAL.limit = nil
		}
	case 219:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1168
		{
			yyVAL.limit = &Limit{Rowcount: yyDollar[2].valExpr}
		}
	case 220:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1172
		{
			yyVAL.limit = &Limit{Offset: yyDollar[2].valExpr, Rowcount: yyDollar[4].valExpr}
		}
	case 221:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1176
		{
			yyVAL.limit = &Limit{Offset: yyDollar[4].valExpr, Rowcount: yyDollar[2].valExpr}
		}
	case 222:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1181
		{
			yyVAL.str = ""
		}
	case 223:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1185
		{
			yyVAL.str = AST_FOR_UPDATE
		}
	case 224:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1189
		{
			if string(yyDollar[3].runes) != SHARE {
				yylex.Error("expecting share")
				return 1
			}
			if string(yyDollar[4].runes) != MODE {
				yylex.Error("expecting mode")
				return 1
			}
			yyVAL.str = AST_SHARE_MODE
		}
	case 225:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1202
		{
			yyVAL.columns = nil
		}
	case 226:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1206
		{
			yyVAL.columns = yyDollar[2].columns
		}
	case 227:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1212
		{
			yyVAL.columns = Columns{&NonStarExpr{Expr: yyDollar[1].colName}}
		}
	case 228:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1216
		{
			yyVAL.columns = append(yyVAL.columns, &NonStarExpr{Expr: yyDollar[3].colName})
		}
	case 229:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1221
		{
			yyVAL.updateExprs = nil
		}
	case 230:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1225
		{
			yyVAL.updateExprs = yyDollar[5].updateExprs
		}
	case 231:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1231
		{
			yyVAL.insRows = yyDollar[2].values
		}
	case 232:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1235
		{
			yyVAL.insRows = yyDollar[1].selStmt
		}
	case 233:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1241
		{
			yyVAL.values = Values{yyDollar[1].rowTuple}
		}
	case 234:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1245
		{
			yyVAL.values = append(yyDollar[1].values, yyDollar[3].rowTuple)
		}
	case 235:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1251
		{
			yyVAL.rowTuple = ValTuple(yyDollar[2].valExprs)
		}
	case 236:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1255
		{
			yyVAL.rowTuple = yyDollar[1].subquery
		}
	case 237:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1261
		{
			yyVAL.updateExprs = UpdateExprs{yyDollar[1].updateExpr}
		}
	case 238:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1265
		{
			yyVAL.updateExprs = append(yyDollar[1].updateExprs, yyDollar[3].updateExpr)
		}
	case 239:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1271
		{
			yyVAL.updateExpr = &UpdateExpr{Name: yyDollar[1].colName, Expr: yyDollar[3].valExpr}
		}
	case 240:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1276
		{
			yyVAL.empty = struct{}{}
		}
	case 241:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1278
		{
			yyVAL.empty = struct{}{}
		}
	case 242:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1281
		{
			yyVAL.empty = struct{}{}
		}
	case 243:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1283
		{
			yyVAL.empty = struct{}{}
		}
	case 244:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1286
		{
			yyVAL.empty = struct{}{}
		}
	case 245:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1288
		{
			yyVAL.empty = struct{}{}
		}
	case 246:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1292
		{
			yyVAL.empty = struct{}{}
		}
	case 247:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1294
		{
			yyVAL.empty = struct{}{}
		}
	case 248:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1296
		{
			yyVAL.empty = struct{}{}
		}
	case 249:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1298
		{
			yyVAL.empty = struct{}{}
		}
	case 250:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1300
		{
			yyVAL.empty = struct{}{}
		}
	case 251:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1303
		{
			yyVAL.empty = struct{}{}
		}
	case 252:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1305
		{
			yyVAL.empty = struct{}{}
		}
	case 253:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1308
		{
			yyVAL.empty = struct{}{}
		}
	case 254:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1310
		{
			yyVAL.empty = struct{}{}
		}
	case 255:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1313
		{
			yyVAL.empty = struct{}{}
		}
	case 256:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1315
		{
			yyVAL.empty = struct{}{}
		}
	case 257:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1318
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 258:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1319
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 259:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1320
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 260:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1321
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 261:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1322
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 262:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1323
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 263:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1324
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 264:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1325
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 265:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1326
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 266:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1327
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 267:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1328
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 268:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1329
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 269:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1330
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 270:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1331
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 271:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1332
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 272:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1333
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 273:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1334
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 274:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1335
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 275:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1336
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 276:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1337
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 277:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1338
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 278:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1339
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 279:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1340
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 280:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1341
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 281:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1342
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 282:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1343
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 283:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1344
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 284:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1345
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 285:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1346
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 286:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1347
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 287:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1348
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 288:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1349
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 289:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1350
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 290:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1351
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 291:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1352
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 292:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1353
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 293:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1354
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 294:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1355
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 295:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1356
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 296:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1357
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 297:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1358
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 298:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1359
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 299:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1360
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 300:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1361
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 301:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1362
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 302:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1363
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 303:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1364
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 304:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1365
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 305:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1366
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 306:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1367
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 307:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1368
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 308:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1369
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 309:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1370
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 310:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1371
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 311:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1372
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 312:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1373
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 313:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1374
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 314:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1375
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 315:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1376
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 316:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1377
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 317:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1378
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 318:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1379
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 319:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1380
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 320:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1381
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 321:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1382
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 322:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1383
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 323:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1384
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 324:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1385
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 325:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1386
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 326:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1387
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 327:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1388
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 328:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1389
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 329:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1390
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 330:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1391
		{
			yyVAL.runes = yyDollar[1].runes
		}
	case 331:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1394
		{
			ForceEOF(yylex)
		}
	}
	goto yystack /* stack new state and value */
}
