import os
import shutil
import subprocess

from utils import run_path, wait_file


def app_cmd(tt_cmd, tmpdir_with_cfg, cmd, input):
    start_cmd = [tt_cmd, *cmd]
    tt_process = subprocess.Popen(
        start_cmd,
        cwd=tmpdir_with_cfg,
        stderr=subprocess.STDOUT,
        stdout=subprocess.PIPE,
        stdin=subprocess.PIPE,
        text=True
    )

    tt_process.stdin.writelines(input)
    tt_process.stdin.close()
    rc = tt_process.wait()
    assert rc == 0
    return tt_process.stdout.readlines()


def test_restart(tt_cmd, tmpdir_with_cfg):
    shutil.copy(os.path.join(os.path.dirname(__file__), "test_app.lua"), tmpdir_with_cfg)
    app_name = "test_app"
    start_output = app_cmd(tt_cmd, tmpdir_with_cfg, ["start", app_name], [])
    assert "Starting an instance" in start_output[0]
    wait_file(os.path.join(tmpdir_with_cfg, run_path, app_name), 'test_app.pid', [])

    try:
        # Test confirmed restart.
        restart_output = app_cmd(tt_cmd, tmpdir_with_cfg, ["restart", app_name], ["y\n"])
        assert "Confirm restart of 'test_app' [y/n]" in restart_output[0]
        assert "has been terminated" in restart_output[0]
        assert "Starting an instance" in restart_output[1]
        wait_file(os.path.join(tmpdir_with_cfg, run_path, app_name), 'test_app.pid', [])

        # Test cancelled restart.
        restart_output = app_cmd(tt_cmd, tmpdir_with_cfg, ["restart", app_name], ["n\n"])
        assert "Confirm restart of 'test_app' [y/n]" in restart_output[0]
        assert "Restart is cancelled" in restart_output[0]
        wait_file(os.path.join(tmpdir_with_cfg, run_path, app_name), 'test_app.pid', [])

    finally:
        app_cmd(tt_cmd, tmpdir_with_cfg, ["stop", app_name], [])


def test_restart_with_auto_yes(tt_cmd, tmpdir_with_cfg):
    shutil.copy(os.path.join(os.path.dirname(__file__), "test_app.lua"), tmpdir_with_cfg)
    app_name = "test_app"
    start_output = app_cmd(tt_cmd, tmpdir_with_cfg, ["start", app_name], [])
    assert "Starting an instance" in start_output[0]
    wait_file(os.path.join(tmpdir_with_cfg, run_path, app_name), 'test_app.pid', [])

    try:
        restart_output = app_cmd(tt_cmd, tmpdir_with_cfg, ["restart", "-y", app_name], [])
        assert "Confirm restart of 'test_app' [y/n]" not in restart_output[0]
        assert "has been terminated" in restart_output[0]
        assert "Starting an instance" in restart_output[1]
        wait_file(os.path.join(tmpdir_with_cfg, run_path, app_name), 'test_app.pid', [])

        restart_output = app_cmd(tt_cmd, tmpdir_with_cfg, ["restart", "--yes", app_name], [])
        assert "Confirm restart of 'test_app' [y/n]" not in restart_output[0]
        assert "has been terminated" in restart_output[0]
        assert "Starting an instance" in restart_output[1]
        wait_file(os.path.join(tmpdir_with_cfg, run_path, app_name), 'test_app.pid', [])

    finally:
        app_cmd(tt_cmd, tmpdir_with_cfg, ["stop", app_name], [])


def test_restart_no_args(tt_cmd, tmpdir):
    test_app_path_src = os.path.join(os.path.dirname(__file__), "multi_app")

    test_app_path = os.path.join(tmpdir, "multi_app")
    shutil.copytree(test_app_path_src, test_app_path)

    start_output = app_cmd(tt_cmd, test_app_path, ["start"], [])
    assert "Starting an instance" in start_output[0]

    try:
        # Test confirmed restart.
        restart_output = app_cmd(tt_cmd, test_app_path, ["restart"], ["y\n"])
        assert "Confirm restart of all instances [y/n]" in restart_output[0]

    finally:
        app_cmd(tt_cmd, test_app_path, ["stop"], [])
