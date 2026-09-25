# .PHONY: entry getin multicurl

entry:
	echo 'entry before getin'

getin:
	echo 'getin after entry'

multicurl:
	chmod +x multicurl.sh && ./multicurl.sh