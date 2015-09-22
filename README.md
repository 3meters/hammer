hammer
======

go-based http client to stress test the proxibase web service

Proxibase has one basic basic test that is designed to simulate real-world client use as closely as possible: /proxibase/test/basic/18_private.js.  When the proxibase server is run with the config flag requestLog=true, It will create a separate log file with just one request per entry.  This program takes that log file as its input.  

In the test, tarzan, jane and mary are playing with patchr in the jungle.  The go program will replay those requests in paralell using go routines.  For each instance, some variables in the test requests are substituted including the logged in user and the location on the globe of the jungle. 

The program runs againts a running proxibase server, and collects various performance statistics that can be used to compare various server, host, and network conditions under simulated load.  
