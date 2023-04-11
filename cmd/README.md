# cmd

# Application 
to manage archived pictures and/or any other files on external devices like hard drives, USB sticks, memory cards, etc.


Usage arguments:
 -h help text
 -u picturesDirName       #update arcpics.json dir files
 -b picturesDirName       #update database about 
 -d databaseDirName       #database dir location other then default ~/.arcpics
 -c databaseDirName label #create label inside of database directory
 -a label                 #list all dirs on USB  with this label
 -s label query           #list specific resources
 -l                       #list all labels
 -p port                  #web port definition
 -w                       #start web - default port 8080

Examples:
-c /media/joe/USB32/Arc-Pics Vacation-2023 #creates label file inside of directory /media/joe/USB32/Arc-Pics
-u /media/joe/USB32/Arc-Pics               #updates arcpics.json dir files inside of directories at /media/joe/USB32/Arc-Pics
-b /media/joe/USB32/Arc-Pics               #updates database ~/.arcpics/Vacation-2023.DB


