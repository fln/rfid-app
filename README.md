rfid-app
========

This is a go library and demo command line application for interfacing with
simple low-frequency 125KHz RFID tag read/writer devices. It works with the same
devices as [rfid_app](https://github.com/merbanan/rfid_app) project.

These devices are detected as Prolific PL2303 USB-serial controllers.

Library
-------

All device comminucation logic is moved to a separate golang package `rfid`.
This library can be used to create custom applications supporting this
particular RFID read/writer device.

[![Documentation](https://godoc.org/github.com/fln/rfid-app/rfid?status.svg)](https://godoc.org/github.com/fln/rfid-app/rfid)

App usage examples
------------------

Read a single card/tag, application will wait until tag is detected:

```sh
$ ./rfid-app 
00000013ec
```

Reading multiple tags in a loop:

```sh
$ ./rfid-app --mode read-loop
00000013ec
00000013ed
```

Checking device model info:

```sh
$ ./rfid-app --mode info
ID card reader & writer
```

Thanks
------

* Damien Bobillot for [device protodol dumps](https://www.triades.net/13-geek/13-serial-protocol-for-a-chinese-rfid-125khz-reader-writer.html).
* Benjamin Larsson for [rfid_app](https://github.com/merbanan/rfid_app).
