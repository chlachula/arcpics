[See general stedolan **jq cookbook**](https://github.com/stedolan/jq/wiki/Cookbook)
***
**print all objects which contains string "inter" in comment**

find Arc-Pics -name "arcpics.json" -exec jq -r '.files[] | select(.comment | . and contains("inter"))' {} \;

**in a directory print about, all comments and selected comment**

cd Arc-Pics/2023/2023_03_23 

jq -r '.about' arcpics.json . 
jq -r '.files[].comment' arcpics.json . 
jq -r '.files[] | select(.comment | . and contains("inter"))' arcpics.json . 
***
[Written with Markdown](https://www.markdownguide.org/basic-syntax/)
