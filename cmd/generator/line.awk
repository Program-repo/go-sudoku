#3 . . |. 1 5 |. . . 
#. 9 2 |6 . . |4 . . 
#1 . . |. . . |. 7 . 
#------+------+------
#. . . |. 6 . |. 9 . 
#. . . |1 . . |. . 7 
#. . 7 |. . 3 |5 4 1 
#------+------+------
#. 6 . |7 . . |. . . 
#. 3 . |. 2 . |. 8 . 
#. . . |. . . |3 . . 
#
#Difficulty: 2.35
# 
#  awk -v snumber=1  -f line.awk input.txt
BEGIN {
	FS="|";
	OFS="";
}
{

prefix = substr($1,1,1)
if (prefix != " " && prefix != "-" && prefix != "T" && prefix != "D") {
    record = $1 $2 $3
    gsub(" ","",record)
	records = records record
}
if (prefix == "D") {
    record = $0
	records = records " #" snumber " " record
}


}
END{
print records	
}
