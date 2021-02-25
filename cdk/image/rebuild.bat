@echo off
cd src\save_metadata

echo Uploading new version of save metadata function
call uploadNewZip.bat ImageStack-docexamplesavemetadata379FAC14-XCDG15QLI6QV

cd ..\save_objectdata
echo Uploading new version of save object data function
call uploadNewZip.bat ImageStack-docexamplesaveobjectdata922528C5-1NGWJHUO7KYZ1

cd ..\create_thumbnail
echo Uploading new version of create thumbnail function
call uploadNewZip.bat ImageStack-docexamplecreatethumbnail163ECDCF-Y7APS4359KKP

cd ..\..

echo Done
