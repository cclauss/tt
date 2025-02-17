========
Examples
========

This file contains various examples of working with tt.

--------
Contents
--------
.. contents::
  :local:

Working with a set of instances
-------------------------------

For example, we want to launch two instances based on one ``init.lua`` file and one
instance based on the ``router.init.lua`` file. In order to do this, we create
a directory called ``demo`` with the content:

``init.lua``:

.. code-block:: lua

    local inst_name = os.getenv('TARANTOOL_INSTANCE_NAME')
    local app_name = os.getenv('TARANTOOL_APP_NAME')

    while true do
        if app_name ~= nil and inst_name ~= nil then
            print(app_name .. ":" .. inst_name)
        else
            print("unknown instance")
        end
        require("fiber").sleep(1)
    end

``router.init.lua``:

.. code-block:: lua

    local inst_name = os.getenv('TARANTOOL_INSTANCE_NAME')
    local app_name = os.getenv('TARANTOOL_APP_NAME')

    while true do
        print("custom init file...")
        if app_name ~= nil and inst_name ~= nil then
            print(app_name .. ":" .. inst_name)
        else
            print("unknown instance")
        end
        require("fiber").sleep(1)
    end

``instances.yml`` (The dot and dash characters in instance names
are reserved for system use.):

.. code-block:: yaml

    router:

    master:

    replica:

Now we can run all instances at once:

.. code-block:: bash

   $ tt start demo
   • Starting an instance [demo:router]...
   • Starting an instance [demo:master]...
   • Starting an instance [demo:replica]...

Or just one of them:

.. code-block:: bash

   $ tt start demo:master
   • Starting an instance [demo:master]...

For starting all instances of environment, run:

.. code-block:: bash

   $ tt start
   • Starting an instance [demo:router]...
   • Starting an instance [demo:master]...
   • Starting an instance [demo:replica]...
   • Starting an instance [demo_single_instance_app]...

Creating Cartridge application
----------------------------------

Create new tt environment, if it is not exist:

.. code-block:: bash

    $ tt init

Create Cartridge application:

.. code-block:: bash

    $ tt create cartridge --name myapp

Build and start the application:

.. code-block:: bash

    $ tt build myapp
    $ tt start myapp

Bootstrap vshard:

.. code-block:: bash

    $ tt cartridge replicasets setup --bootstrap-vshard --name myapp --run-dir ./var/run/myapp/

Now open http://localhost:8081/ and see your application's Admin Web UI.

Working with application templates
----------------------------------

For example, we want to create an application template. In order to do this, create a directory for the template:

.. code-block:: bash

    $ mkdir -p ./templates/simple

with the content:

``init.lua.tt.template``:

.. code-block:: lua

    local app_name = {{.name}}
    local login = {{.user_name}}

    require("fiber").sleep(1)

``MANIFEST.yaml``:

.. code-block:: yaml

    description: Simple app
    vars:
        - prompt: User name
          name: user_name
          default: admin
          re: ^\w+$

``init.lua.tt.template`` in this example contains an application code. After instantiation, ``.tt.template`` suffix is removed from the file name.

Create ``./tt.yaml`` and add templates search path to it:

.. code-block:: yaml

    tt:
        templates:
            - path: ./templates

Here is how the current directory structure looks like::

    ./
    ├── tt.yaml
    └── templates
        └── simple
            ├── init.lua.tt.template
            └── MANIFEST.yaml

Directory name ``simple`` can now be used as template name in create command.
Create an application from the ``simple`` template and type ``user1`` in ``User name`` prompt:

.. code-block:: bash

   $ tt create simple --name simple_app
   • Creating application in <current_directory>/simple_app
   • Using template from <current_directory>/templates/simple
   User name (default: admin): user1

Your application will appear in the ``simple_app`` directory with the following content::

    simple_app/
    ├── Dockerfile.build.tt
    └── init.lua

Instantiated ``init.lua`` content:

.. code-block:: lua

    local app_name = simple_app
    local login = user1

    require("fiber").sleep(1)

Packing environments
--------------------

For example, we want to pack a single application. Here is the content of the sample application::

    single_environment/
    ├── tt.yaml
    └── init.lua

``tt.yaml``:

.. code-block:: yaml

    tt:
        app:

For packing it into tarball, call:

.. code-block:: bash

   $ tt pack tgz
      • Apps to pack: single_environment
      • Generating new tt.yaml for the new package.
      • Creating tarball.
      • Bundle is packed successfully to /Users/dev/tt_demo/single_environment/single_environment_0.1.0.0.tar.gz.

The result directory structure::

      unpacked_dir/
      ├── tt.yaml
      ├── single_environment
      │   └── init.lua
      ├── env
      │   ├── bin
      │   └── modules
      ├── instances_enabled
      │   └── single_environment -> ../single_environment
      └── var
          ├── lib
          ├── log
          └── run

Example of packing a multi-app environment. The source tree::

     bundle/
     ├── tt.yaml
     ├── env
     │   ├── bin
     │   │   ├── tt
     │   │   └── tarantool
     │   └── modules
     ├── myapp
     │   ├── Dockerfile.build.cartridge
     │   ├── Dockerfile.cartridge
     │   ├── README.md
     │   ├── app
     │   ├── bin
     │   ├── deps.sh
     │   ├── failover.yml
     │   ├── init.lua
     │   ├── instances.yml
     │   ├── myapp-scm-1.rockspec
     │   ├── pack-cache-config.yml
     │   ├── package-deps.txt
     │   ├── replicasets.yml
     │   ├── stateboard.init.lua
     │   ├── systemd-unit-params.yml
     │   ├── tt.yaml
     │   ├── test
     │   └── tmp
     ├── myapp2
     │   ├── app.lua
     │   ├── data
     │   ├── etc
     │   ├── myapp2
     │   ├── queue
     │   ├── queue1.lua
     │   └── queue2.lua
     ├── myapp3.lua
     ├── app4.lua
     ├── instances_enabled
     │   ├── app1 -> ../myapp
     │   ├── app2 -> ../myapp2
     │   ├── app3.lua -> ../myapp3.lua
     │   ├── app4.lua -> /Users/dev/tt_demo/bundle1/app4.lua
     │   └── app5.lua -> ../myapp3.lua
     └── var
         ├── lib
         ├── log
         └── run

``tt.yaml``:

.. code-block:: yaml

    tt:
      modules:
        directory: env/modules
      app:
        instances_enabled: instances_enabled
        run_dir: var/run
        log_dir: var/log
        log_maxsize: 1
        log_maxage: 1
        log_maxbackups: 1
        restart_on_failure: true
        wal_dir: var/lib
        vinyl_dir: var/lib
        memtx_dir: var/lib
        bin_dir: env/bin

Pay attention, that all absolute symlinks from `instances_enabled` will be resolved, all sources will be copied
to the result package and the final instances_enabled directory will contain only relative links.

For packing deb package call:

.. code-block:: bash

   $ tt pack deb --name dev_bundle --version 1.0.0
   • A root for package is located in: /var/folders/c6/jv1r5h211dn1280d75pmdqy80000gp/T/2166098848
      • Apps to pack: app1 app2 app3 app4 app5

   myapp scm-1 is now installed in /var/folders/c6/jv1r5h211dn1280d75pmdqy80000gp/T/tt_pack4173588242/myapp/.rocks

      • myapp rocks are built successfully
      • Generating new tt.yaml for the new package
      • Initialize the app directory for prefix: data/usr/share/tarantool/bundle
      • Create data tgz
      • Created control in /var/folders/***/control_dir
      • Created result DEB package: /var/folders/***/T/tt_pack4173588242

Now the result package may be distributed and installed using dpkg command.
The package will be installed in /usr/share/tarantool/package_name directory.

Working with tt daemon (experimental)
-------------------------------------

``tt daemon`` module is used to manage ``tt`` running
on the background on a given machine. This way instances
can be operated remotely.
Daemon can be configured with ``tt_daemon.yaml`` config.

You can manage TT daemon with following commands:

* ``tt daemon start`` - launch of a daemon
* ``tt daemon stop`` - terminate of the daemon
* ``tt daemon status`` - get daemon status
* ``tt daemon restart`` - daemon restart

Work scenario:

First, TT daemon should be started on the server side:

.. code-block:: bash

   $ tt daemon start
   • Starting tt daemon...

After daemon launch you can check its status on the server side:

.. code-block:: bash

   $ tt daemon status
   • RUNNING. PID: 6189.

To send request to daemon you can use CURL. In this example the
client sends a request to start ``test_app`` instance on the server side.
Note: directory ``test_app`` (or file ``test_app.lua``) exists
on the server side.

.. code-block:: bash

   $ curl --header "Content-Type: application/json" --request POST \
   --data '{"command_name":"start", "params":["test_app"]}' \
   http://127.0.0.1:1024/tarantool
   {"res":"   • Starting an instance [test_app]...\n"}

Below is an example of running a command with flags.

Flag with argument:

.. code-block:: bash

   $ curl --header "Content-Type: application/json" --request POST \
   --data '{"command_name":"version", "params":["-L", "/path/to/local/dir"]}' \
   http://127.0.0.1:1024/tarantool
   {"res":"Tarantool CLI version 0.1.0, darwin/amd64. commit: bf83f33\n"}

Flag without argument:

.. code-block:: bash

   $ curl --header "Content-Type: application/json" --request POST \
   --data '{"command_name":"version", "params":["-V"]}' \
   http://127.0.0.1:1024/tarantool
   {"res":"Tarantool CLI version 0.1.0, darwin/amd64. commit: bf83f33\n
    • Tarantool executable found: '/usr/local/bin/tarantool'\n"}

Transition from tarantoolctl to tt
----------------------------------

System-wide configuration
~~~~~~~~~~~~~~~~~~~~~~~~~
``tt`` packages come with a system-wide environment configuration which supports ``tarantoolctl``
configuration defaults. So, after installation from repository ``tt`` can be used along with
``tarantoolctl`` for managing applications instances. Here is an example:

.. code-block::

    $ sudo tt instances
    List of enabled applications:
    • example

    $ tarantoolctl start example
    Starting instance example...
    Forwarding to 'systemctl start tarantool@example'

    $ tarantoolctl status example
    Forwarding to 'systemctl status tarantool@example'
    ● tarantool@example.service - Tarantool Database Server
        Loaded: loaded (/lib/systemd/system/tarantool@.service; enabled; vendor preset: enabled)
        Active: active (running)
        Docs: man:tarantool(1)
        Main PID: 6698 (tarantool)
    . . .

    $ sudo tt status
    • example: RUNNING. PID: 6698.

    $ sudo tt connect example
    • Connecting to the instance...
    • Connected to /var/run/tarantool/example.control

    /var/run/tarantool/example.control> 

    $ sudo tt stop example
    • The Instance example (PID = 6698) has been terminated.

    $ tarantoolctl status example
    Forwarding to 'systemctl status tarantool@example'
    ○ tarantool@example.service - Tarantool Database Server
        Loaded: loaded (/lib/systemd/system/tarantool@.service; enabled; vendor preset: enabled)
        Active: inactive (dead)

Local configuration
~~~~~~~~~~~~~~~~~~~

Suppose we have a set of Tarantool instances managed by ``tarantoolctl`` utility in local directory.
In order for the tt to work with these instances run ``tt init`` command in the directory where
``tarantoolctl`` configuration file (``.tarantoolctl``) is located. For example:

.. code-block:: bash

    $ cat .tarantoolctl 
    default_cfg = {
        pid_file  = "./run/tarantool",
        wal_dir   = "./lib/tarantool",
        memtx_dir = "./lib/tarantool",
        vinyl_dir = "./lib/tarantool",
        log       = "./log/tarantool",
        language  = "Lua",
    }
    instance_dir = "./instances.enabled"

    $ tt init
    • Found existing config '.tarantoolctl'
    • Environment config is written to 'tt.yaml'

``tt init`` takes directories paths from the existing ``tarantoolctl`` config. Generated ``tt.yaml``
will look like:

.. code-block:: yaml

    tt:
    modules:
        directory: modules
    app:
        run_dir: ./run/tarantool
        log_dir: ./log/tarantool
        log_maxsize: 100
        log_maxage: 8
        log_maxbackups: 10
        restart_on_failure: false
        wal_dir: ./lib/tarantool
        memtx_dir: ./lib/tarantool
        vinyl_dir: ./lib/tarantool
        bin_dir: bin
        inc_dir: include
        instances_enabled: ./instances.enabled
        tarantoolctl_layout: true
    ee:
        credential_path: ""
    templates:
    - path: templates
    repo:
        rocks: ""
        distfiles: distfiles

After that we can use ``tt`` for managing Tarantool instances, checkpoint files and modules. Most of
``tt`` commands correspond to the same in ``tarantoolctl``:

::

    $ tt start app1
    • Starting an instance [app1]...

    $ tt status app1
    • app1: RUNNING. PID: 33837.

    $ tt stop app1
    • The Instance app1 (PID = 33837) has been terminated.

    $ tt check app1
    • Result of check: syntax of file '/home/user/instances.enabled/app1.lua' is OK

Commands difference:
~~~~~~~~~~~~~~~~~~~~

``tarantoolctl enter/connect/eval`` functionality is covered by ``tt connect`` command.

``tarantoolctl``:

::

    $ tarantoolctl enter app1
    connected to unix/:./run/tarantool/app1.control
    unix/:./run/tarantool/app1.control>

    $ tarantoolctl connect localhost:3301
    connected to localhost:3301
    localhost:3301>

    $ tarantoolctl eval app1 eval.lua
    connected to unix/:./run/tarantool/app1.control
    ---
    - 42
    ...

    $ cat eval.lua | tarantoolctl eval app1
    connected to unix/:./run/tarantool/app1.control
    ---
    - 42
    ...

``tt`` analog:

::

    $ tt connect app1
    • Connecting to the instance...
    • Connected to /home/user/run/tarantool/app1/app1.control

    /home/user/run/tarantool/app1/app1.control>

    $ tt connect localhost:3301
    • Connecting to the instance...
    • Connected to localhost:3301

    localhost:3301>

    $ tt connect app1 -f eval.lua
    ---
    - 42
    ...

    $ cat eval.lua | tt connect app1 -f -
    ---
    - 42
    ...

Transition from Cartridge CLI to tt
-----------------------------------
To start using ``tt`` for an existing Cartridge application run ``tt init`` command in
application's directory:

::

   $ tt init
   • Found existing config '.cartridge.yml'
   • Environment config is written to 'tt.yaml'

``tt init`` searches for ``.cartridge.yml`` and uses configured paths from it in ``tt.yaml``. After
``tt.yaml`` config is generated, ``tt`` can be used to manage cartridge application. Most of
``cartridge`` commands have their counterparts in ``tt``: build, start, stop, status, etc.:

::

    $ tt start
    • Starting an instance [app:s1-master]...
    • Starting an instance [app:s1-replica]...
    • Starting an instance [app:s2-master]...
    • Starting an instance [app:s2-replica]...
    • Starting an instance [app:stateboard]...
    • Starting an instance [app:router]...
    $ tt status
    • app:router: RUNNING. PID: 13188.
    • app:s1-master: RUNNING. PID: 13177.
    • app:s1-replica: RUNNING. PID: 13178.
    • app:s2-master: RUNNING. PID: 13179.
    • app:s2-replica: RUNNING. PID: 13180.
    • app:stateboard: RUNNING. PID: 13182.

Commands difference:
~~~~~~~~~~~~~~~~~~~~

* Some of cartridge application management commands are subcommands of ``tt cartridge``:

::

    USAGE
      tt cartridge [flags] <command> [command flags]

    COMMANDS
      admin       Call admin function
      bench       Util for running benchmarks for Tarantool
      failover    Manage application failover
      repair      Patch cluster configuration files
      replicasets Manage application replica sets

* ``cartridge enter`` and ``cartridge connect`` functionality is covered by ``tt connect``:

::

    $ tt connect app:router
    • Connecting to the instance...
    • Connected to /home/user/app/var/run/app/router/router.control

    /home/user/app/var/run/app/router/router.control>

    $ tt connect localhost:3302
    • Connecting to the instance...
    • Connected to localhost:3302

    localhost:3302>

* ``cartridge log`` and ``cartridge pack docker`` functionality is not supported in ``tt`` yet.
* Shell autocompletion scripts generation command ``cartridge gen completion`` is ``tt completion``
  in ``tt``.
