import os
import shutil
import subprocess
import tempfile


def test_build_no_options(tt_cmd, tmpdir):
    app_dir = shutil.copytree(os.path.join(os.path.dirname(__file__), "apps/app1"),
                              os.path.join(tmpdir, "app1"))

    buid_cmd = [tt_cmd, "build"]
    tt_process = subprocess.Popen(
        buid_cmd,
        cwd=app_dir,
        stderr=subprocess.STDOUT,
        stdout=subprocess.PIPE,
        stdin=subprocess.PIPE,
        text=True
    )
    tt_process.stdin.close()
    tt_process.wait()
    assert tt_process.returncode == 0

    assert os.path.exists(os.path.join(app_dir, ".rocks", "share", "tarantool", "checks.lua"))
    assert os.path.exists(os.path.join(app_dir, ".rocks", "share", "tarantool", "rocks"))


def test_build_with_tt_hooks(tt_cmd, tmpdir):
    app_dir = shutil.copytree(os.path.join(os.path.dirname(__file__), "apps/app1"),
                              os.path.join(tmpdir, "app1"))
    shutil.copytree(os.path.join(os.path.dirname(__file__), "apps/tt_hooks"),
                    os.path.join(tmpdir, "app1"), dirs_exist_ok=True)

    buid_cmd = [tt_cmd, "build"]
    tt_process = subprocess.Popen(
        buid_cmd,
        cwd=app_dir,
        stderr=subprocess.STDOUT,
        stdout=subprocess.PIPE,
        stdin=subprocess.PIPE,
        text=True
    )
    tt_process.stdin.close()
    tt_process.wait()
    assert tt_process.returncode == 0

    assert os.path.exists(os.path.join(app_dir, ".rocks", "share", "tarantool", "checks.lua"))
    assert os.path.exists(os.path.join(app_dir, ".rocks", "share", "tarantool", "rocks"))
    assert os.path.exists(os.path.join(app_dir, "tt-pre-build-invoked"))
    assert os.path.exists(os.path.join(app_dir, "tt-post-build-invoked"))


def test_build_with_cartridge_hooks(tt_cmd, tmpdir):
    app_dir = shutil.copytree(os.path.join(os.path.dirname(__file__), "apps/app1"),
                              os.path.join(tmpdir, "app1"))
    shutil.copytree(os.path.join(os.path.dirname(__file__), "apps/cartridge_hooks"),
                    os.path.join(tmpdir, "app1"), dirs_exist_ok=True)

    buid_cmd = [tt_cmd, "build"]
    tt_process = subprocess.Popen(
        buid_cmd,
        cwd=app_dir,
        stderr=subprocess.STDOUT,
        stdout=subprocess.PIPE,
        stdin=subprocess.PIPE,
        text=True
    )
    tt_process.stdin.close()
    tt_process.wait()
    assert tt_process.returncode == 0

    assert os.path.exists(os.path.join(app_dir, ".rocks", "share", "tarantool", "checks.lua"))
    assert os.path.exists(os.path.join(app_dir, ".rocks", "share", "tarantool", "rocks"))
    assert os.path.exists(os.path.join(app_dir, "cartridge-pre-build-invoked"))
    assert os.path.exists(os.path.join(app_dir, "cartridge-post-build-invoked"))


def test_build_app_name_set(tt_cmd, tmpdir):
    app_dir = shutil.copytree(os.path.join(os.path.dirname(__file__), "apps/app1"),
                              os.path.join(tmpdir, "app1"))

    buid_cmd = [tt_cmd, "build", "app1"]
    tt_process = subprocess.Popen(
        buid_cmd,
        cwd=tmpdir,
        stderr=subprocess.STDOUT,
        stdout=subprocess.PIPE,
        stdin=subprocess.PIPE,
        text=True
    )
    tt_process.stdin.close()
    tt_process.wait()
    assert tt_process.returncode == 0

    assert os.path.exists(os.path.join(app_dir, ".rocks", "share", "tarantool", "checks.lua"))
    assert os.path.exists(os.path.join(app_dir, ".rocks", "share", "tarantool", "rocks"))


def test_build_absolute_path(tt_cmd, tmpdir):
    app_dir = shutil.copytree(os.path.join(os.path.dirname(__file__), "apps/app1"),
                              os.path.join(tmpdir, "app1"))

    with tempfile.TemporaryDirectory() as tmpWorkDir:
        buid_cmd = [tt_cmd, "build", app_dir]
        tt_process = subprocess.Popen(
            buid_cmd,
            cwd=tmpWorkDir,
            stderr=subprocess.STDOUT,
            stdout=subprocess.PIPE,
            stdin=subprocess.PIPE,
            text=True
        )
        tt_process.stdin.close()
        tt_process.wait()
        assert tt_process.returncode == 0

        assert os.path.exists(os.path.join(app_dir, ".rocks", "share", "tarantool", "checks.lua"))
        assert os.path.exists(os.path.join(app_dir, ".rocks", "share", "tarantool", "rocks"))


def test_build_missing_rockspec(tt_cmd, tmpdir):
    buid_cmd = [tt_cmd, "build"]
    tt_process = subprocess.Popen(
        buid_cmd,
        cwd=tmpdir,
        stderr=subprocess.STDOUT,
        stdout=subprocess.PIPE,
        stdin=subprocess.PIPE,
        text=True
    )
    tt_process.stdin.close()
    tt_process.wait()
    assert tt_process.returncode == 1

    tt_process.stdout.readline()  # Skip empty line.
    assert tt_process.stdout.readline().find(
        "please specify a rockspec to use on current directory") != -1


def test_build_missing_app_dir(tt_cmd, tmpdir):
    buid_cmd = [tt_cmd, "build", "app1"]
    tt_process = subprocess.Popen(
        buid_cmd,
        cwd=tmpdir,
        stderr=subprocess.STDOUT,
        stdout=subprocess.PIPE,
        stdin=subprocess.PIPE,
        text=True
    )
    tt_process.stdin.close()
    tt_process.wait()
    assert tt_process.returncode == 1

    assert tt_process.stdout.readline().find("app1: no such file or directory") != -1


def test_build_multiple_paths(tt_cmd, tmpdir):
    buid_cmd = [tt_cmd, "build", "app1", "app2"]
    tt_process = subprocess.Popen(
        buid_cmd,
        cwd=tmpdir,
        stderr=subprocess.STDOUT,
        stdout=subprocess.PIPE,
        stdin=subprocess.PIPE,
        text=True
    )
    tt_process.stdin.close()
    tt_process.wait()
    assert tt_process.returncode == 1

    assert tt_process.stdout.readline().find("Error: accepts at most 1 arg(s), received 2") != -1


def test_build_spec_file_set(tt_cmd, tmpdir):
    app_dir = shutil.copytree(os.path.join(os.path.dirname(__file__), "apps/app1"),
                              os.path.join(tmpdir, "app1"))

    buid_cmd = [tt_cmd, "build", "app1", "--spec", "app1-scm-1.rockspec"]
    tt_process = subprocess.Popen(
        buid_cmd,
        cwd=tmpdir,
        stderr=subprocess.STDOUT,
        stdout=subprocess.PIPE,
        stdin=subprocess.PIPE,
        text=True
    )
    tt_process.stdin.close()
    tt_process.wait()
    assert tt_process.returncode == 0

    assert os.path.exists(os.path.join(app_dir, ".rocks", "share", "tarantool", "checks.lua"))
    assert os.path.exists(os.path.join(app_dir, ".rocks", "share", "tarantool", "rocks"))
    assert os.path.exists(os.path.join(app_dir, ".rocks", "share", "tarantool", "metrics"))
    assert os.path.exists(os.path.join(app_dir, ".rocks", "share", "tarantool", "cartridge"))
