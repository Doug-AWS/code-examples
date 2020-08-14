<?php
$s="abc xyz";

$ar=explode(" ", $s);

$len=count($ar);

for ($i=0; $i<$len; $i++) {
    echo "$ar[$i] \n";
}

echo "\n";

for ($i=$len-1; $i>=0; $i--) {
    echo "$ar[$i] \n";
}

echo "\n";

$a = '1';
$b = &$a;
$b = "2$b";
echo $a.", ".$b;
?>