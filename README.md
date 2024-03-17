# Finding Registry

This service keeps track of what assets have findings reported on it

## Finding

### Report Locator

A finding is reported on a singular locator. A locator is *map* leading to where the finding was found and also represents the scope of the finding.
A locator is a combination of a type and a value where a format is implied based on the type. For example

* IPv4, 192.168.0.1
* TCP, 192.168.0.1:443
* TCP, yeet.com:443
* UDP, [0:::0]:443

### Implied Report Locators

Whenever a finding is reported on a `report locator`, a set of `implied locators` are calculated based on the `report locator`. For example `192.168.0.1:443` of type `TCP` would have the following implied list of locators

* `192.168.0.1:443` of type `TCP`
* `192.168.0.1` of type `IPv4`

`implied locators` are only calculated based on information that is immediately available in the `report locator`. This means that the following list can not be caluclated.

* Hostname resolution to IP address
* IP address link to MAC address

