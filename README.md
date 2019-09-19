# MLM - Mesos Log Monitor [![Actions Status](https://github.com/erikbozic/mlm/workflows/go-build/badge.svg)](https://github.com/erikbozic/mlm/actions)
**Work in progress!**    
Command line tool for monitoring mesos task logs

# Usage
Run binary with flag `-m` and pass in the http url for mesos master.  
`` ./mlm -m http://localhost:5050 ``  
This will get all tasks known to master and let you specify tasks which you want to monitor logs from.

After first usage the mesos master url will saved in a configuration file for next time.

# Commands
 - `:b`: issue this command to stop listening and return you to the task selection  
 - `:q`: issue this command to stop listening and quit the program  
 - `:f {filterString}` issue this command to start filtering log messages on all listeners using the provided filter string
  
