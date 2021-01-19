# echo args
for f in "$@" ; do
    if [ "$f" == "*_test.go" ]
    then
       echo $f
    fi
done
