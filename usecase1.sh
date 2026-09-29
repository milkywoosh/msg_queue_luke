# add curl [POST, GET]



curl -X POST "http://localhost:8005/api/v1/counting-item-per-productcode" \
  -F "file-upload1=@C:\csv_rand\data-person.csv;type=text/csv" \
  -F "note=any note to write"


curl -X GET "http://localhost:8005/api/v1/product/count/:product_code" \

curl -v -X GET "http://localhost:8005/api/v1/product/count/PRD-001"

URL-encoding is only needed for characters that have special meaning in a URL or aren't URL-safe, such as:
--------------------------------
Character	    |Encoded	Needed?
----------------|-------------------
- (dash)	    | -	    No, safe as-is
_ (underscore)	| _	    No, safe as-is
/ (slash)	    | %2F	Yes, if it's part of the value (not a path separator)
(space)	        | %20	Yes
#	            | %23	Yes
&	            | %26	Yes