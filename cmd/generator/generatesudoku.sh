#!/bin/bash
echo "Start = " `date`
truncate -s 0 sudokus.txt
i=1
while [[ $i -le 1000 ]] ; do
    # Generate 
    ./generator -diff 1 > sudoku1.txt 
    # create line
    awk -v snumber="$i" -f line.awk sudoku1.txt >> sudokus.txt
    # end
  (( i += 1 ))
done
    echo "$i"
echo "End   = " `date`
