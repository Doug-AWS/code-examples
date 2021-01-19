# echo args

TestGoFile () {
    if [ "$1" != "" ]
    then
	path="$(dirname $1)"
        file="$(basename $1)"
	pushd $path

	declare RESULT=(`go test`)  # (..) = array
	
	if [ "${RESULT[0]}" == "PASS" ]
	then
	    echo 0
	else
	    echo 1
	fi
	popd
    fi
}

for f in $@ ; do
    # Do any end with "_test.go"?
    [[ $f =~ ^[a-zA-Z/]*_test.go$ ]] && TestGoFile "$f"
done
