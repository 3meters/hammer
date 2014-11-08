
Running locally on production server
====================================

Config:
{
  "Host": "https://localhost:8443",
  "Signin": {
    "Email": "admin",
    "Password": "admin",
    "InstallId": "1"
  },
  "Seed": "625195",
  "Hammers": 5,
  "Seconds": 120,
  "MaxProcs": 1,
  "RequestPath": "request.log",
  "Log": false
}

Results: 

Runs: 379
Requests: 14525
Errors: 0
Fail Rate: 0.00
KBytes per second: 153
Requests per second: 121
Min time: 37
Max time: 7998
Mean time: 412
Median time: 193


Result2: 

Runs: 426
Requests: 16273
Errors: 0
Fail Rate: 0.00
KBytes per second: 172
Requests per second: 135
Min time: 33
Max time: 7237
Mean time: 367
Median time: 172


Results3:

Runs: 406
Requests: 15549
Errors: 0
Fail Rate: 0.00
KBytes per second: 164
Requests per second: 129
Min time: 26
Max time: 7270
Mean time: 385
Median time: 183

Result 4:
Runs: 403
Requests: 15371
Errors: 0
Fail Rate: 0.00
KBytes per second: 162
Requests per second: 128
Min time: 26
Max time: 8160
Mean time: 389
Median time: 185


Joyent 3.5gb,
-----------------------------------------

{
  "Host": "https://api.aircandi.com:8443",
  "Signin": {
    "Email": "admin",
    "Password": "admin",
    "InstallId": "1"
  },
  "Seed": "625195",
  "Hammers": 5,
  "Seconds": 120,
  "RequestPath": "request.log",
  "Log": false
}

Runs: 203
Requests: 7835
Errors: 0
Fail Rate: 0.00
KBytes per second: 84
Requests per second: 65
Min time: 28
Max time: 8669
Mean time: 302
Median time: 127

Runs: 201
Requests: 7737
Errors: 0
Fail Rate: 0.00
KBytes per second: 83
Requests per second: 64
Min time: 27
Max time: 8727
Mean time: 308
Median time: 129



From Amazon, no CloudFlare:  IP: 165.225.151.254
=================================================


{
  "Host": "https://api.aircandi.com:444",
  "Signin": {
    "Email": "admin",
    "Password": "admin",
    "InstallId": "1"
  },
  "Seed": "625195",
  "Hammers": 5,
  "Seconds": 120,
  "RequestPath": "request.log",
  "Log": false
}

Results1:

Runs: 135
Requests: 5227
Errors: 0
Fail Rate: 0.00
KBytes per second: 55
Requests per second: 43
Min time: 277
Max time: 8573
Mean time: 539
Median time: 375

Results2: 

Runs: 137
Requests: 5334
Errors: 0
Fail Rate: 0.00
KBytes per second: 56
Requests per second: 44
Min time: 269
Max time: 7298
Mean time: 523
Median time: 368


Results3: 

Runs: 144
Requests: 5562
Errors: 0
Fail Rate: 0.00
KBytes per second: 58
Requests per second: 46
Min time: 275
Max time: 7242
Mean time: 484
Median time: 356


From Amazon, Using Cloudfare free, IP:  104.28.29.126
===================================================

Config:
{
  "Host": "https://api.aircandi.com:8443",
  "Signin": {
    "Email": "admin",
    "Password": "admin",
    "InstallId": "1"
  },
  "Seed": "625195",
  "Hammers": 5,
  "Seconds": 120,
  "RequestPath": "request.log",
  "Log": false
}

Results:

Runs: 191
Requests: 7388
Errors: 0
Fail Rate: 0.00
KBytes per second: 78
Requests per second: 61
Min time: 274
Max time: 8305
Mean time: 805
Median time: 704

Results:

Runs: 197
Requests: 7618
Errors: 0
Fail Rate: 0.00
KBytes per second: 81
Requests per second: 63
Min time: 278
Max time: 10026
Mean time: 781
Median time: 701


Runs: 199
Requests: 7676
Errors: 0
Fail Rate: 0.00
KBytes per second: 81
Requests per second: 63
Min time: 273
Max time: 8104
Mean time: 774
Median time: 701

Runs: 200
Requests: 7707
Errors: 0
Fail Rate: 0.00
KBytes per second: 82
Requests per second: 64
Min time: 271
Max time: 7101
Mean time: 771
Median time: 700



From Amazon, Using Cloudfare Pro, ($20 / mth) Default settings, IP:  108.162.203.88
===================================================================================

Config:
{
  "Host": "https://api.aircandi.com:8443",
  "Signin": {
    "Email": "admin",
    "Password": "admin",
    "InstallId": "1"
  },
  "Seed": "625195",
  "Hammers": 5,
  "Seconds": 120,
  "RequestPath": "request.log",
  "Log": false
}

Results: 

Runs: 156
Requests: 5973
Errors: 0
Fail Rate: 0.00
KBytes per second: 63
Requests per second: 49
Min time: 347
Max time: 9260
Mean time: 1000
Median time: 926

Runs: 155
Requests: 6024
Errors: 0
Fail Rate: 0.00
KBytes per second: 64
Requests per second: 50
Min time: 355
Max time: 8328
Mean time: 992
Median time: 936

Runs: 157
Requests: 6083
Errors: 0
Fail Rate: 0.00
KBytes per second: 64
Requests per second: 50
Min time: 354
Max time: 6897
Mean time: 981
Median time: 926


From Amazon, Cloudflair Paused, IP:  165.225.151.254
=================================================

Runs: 142
Requests: 5454
Errors: 0
Fail Rate: 0.00
KBytes per second: 57
Requests per second: 45
Min time: 276
Max time: 7582
Mean time: 502
Median time: 356

Runs: 143
Requests: 5516
Errors: 0
Fail Rate: 0.00
KBytes per second: 58
Requests per second: 45
Min time: 272
Max time: 8273
Mean time: 491
Median time: 354

Runs: 144
Requests: 5592
Errors: 0
Fail Rate: 0.00
KBytes per second: 59
Requests per second: 46
Min time: 270
Max time: 5835
Mean time: 477
Median time: 352


From Amazon, Cloudflair Pro Resumed, Security set essentiall off, IP:  108.162.204.152
==========================
Runs: 155
Requests: 6021
Errors: 0
Fail Rate: 0.00
KBytes per second: 63
Requests per second: 50
Min time: 351
Max time: 10861
Mean time: 992
Median time: 929

Runs: 156
Requests: 6038
Errors: 0
Fail Rate: 0.00
KBytes per second: 64
Requests per second: 50
Min time: 360
Max time: 6331
Mean time: 988
Median time: 927

Runs: 159
Requests: 6148
Errors: 0
Fail Rate: 0.00
KBytes per second: 65
Requests per second: 51
Min time: 358
Max time: 8828
Mean time: 971
Median time: 921


Proxstage local host, Joyent 1.75 GB box
========================================
Results: 

Runs: 206
Requests: 7973
Errors: 0
Fail Rate: 0.00
KBytes per second: 85
Requests per second: 66
Min time: 30
Max time: 6081
Mean time: 239
Median time: 109

Runs: 206
Requests: 7931
Errors: 0
Fail Rate: 0.00
KBytes per second: 85
Requests per second: 66
Min time: 28
Max time: 7932
Mean time: 246
Median time: 110

Proxstage local host, Joyent 4 GB box
========================================
{
  "Host": "https://api.aircandi.com:8443",
  "Signin": {
    "Email": "admin",
    "Password": "admin",
    "InstallId": "1"
  },
  "Seed": "625195",
  "Hammers": 5,
  "Seconds": 120,
  "RequestPath": "request.log",
  "Log": false
}

Results: 

Runs: 142
Requests: 5450
Errors: 0
Fail Rate: 0.00
KBytes per second: 58
Requests per second: 45
Min time: 28
Max time: 7364
Mean time: 254
Median time: 114

Runs: 142
Requests: 5552
Errors: 0
Fail Rate: 0.00
KBytes per second: 59
Requests per second: 46
Min time: 28
Max time: 6964
Mean time: 224
Median time: 106

Runs: 141
Requests: 5461
Errors: 0
Fail Rate: 0.00
KBytes per second: 58
Requests per second: 45
Min time: 28
Max time: 5914
Mean time: 249
Median time: 109




From Amazon to Joyent 4gb, no cloudflare
-----------------------------------------

Results: 

Runs: 133
Requests: 5125
Errors: 0
Fail Rate: 0.00
KBytes per second: 53
Requests per second: 42
Min time: 282
Max time: 3607
Mean time: 561
Median time: 399

Results: 

Runs: 137
Requests: 5302
Errors: 0
Fail Rate: 0.00
KBytes per second: 55
Requests per second: 44
Min time: 285
Max time: 2696
Mean time: 525
Median time: 380


Home mac, local, mongo 2.4.11
------------------

Result: 

Runs: 934
Requests: 35633
Errors: 0
Fail Rate: 0.00
KBytes per second: 378
Requests per second: 296
Min time: 16
Max time: 1455
Mean time: 168
Median time: 83

Runs: 987
Requests: 37626
Errors: 0
Fail Rate: 0.00
KBytes per second: 399
Requests per second: 313
Min time: 18
Max time: 1411
Mean time: 159
Median time: 75

(60 seconds, not 120)
Runs: 475
Requests: 18124
Errors: 0
Fail Rate: 0.00
KBytes per second: 385
Requests per second: 302
Min time: 16
Max time: 1756
Mean time: 165
Median time: 80

(60 seconds, db, not clean, npm update)
Runs: 261
Requests: 10059
Errors: 0
Fail Rate: 0.00
KBytes per second: 214
Requests per second: 167
Min time: 18
Max time: 3129
Mean time: 298
Median time: 156

(db clean)
Runs: 474
Requests: 18071
Errors: 0
Fail Rate: 0.00
KBytes per second: 382
Requests per second: 301
Min time: 21
Max time: 1673
Mean time: 165
Median time: 83



