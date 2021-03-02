TestGoFile () {
    if [ "$1" == "" ]
    then
       return 0
    fi

    pushd $1 > /dev/nul 2>&1

    declare RESULT=(`go test`)  # (..) = array

#    echo RESULT in TestGoFile: $RESULT
    
    if [ "${RESULT[0]}" == "PASS" ]
    then
      echo 0
    else
      echo 1
    fi

    popd > /dev/nul 2>&1
}

for f in $@ ; do
#    echo Looking at $f
    # Do any end with "_test.go"?
    path="$(dirname $f)"
    file="$(basename $f)"

#    echo Path is $path
#    echo filename is $file

    # If it's a go test file
    # test it
    case $file in
        *[a-zA-Z]*_test.go)
#    	    echo It IS a go test file
            TestGoFile "$path"
	    ;;
        *)
#    	    echo It is NOT a go test file
            echo 0
    esac
done
