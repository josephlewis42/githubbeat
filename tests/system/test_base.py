from githubbeat import BaseTest

import os


class Test(BaseTest):

    def test_base(self):
        """
        Basic test with exiting Githubbeat normally
        """
        self.render_config_template(
            path=os.path.abspath(self.working_dir) + "/log/*"
        )

        githubbeat_proc = self.start_beat()
        self.wait_until(lambda: self.log_contains("githubbeat is running"))
        exit_code = githubbeat_proc.kill_and_wait()
        assert exit_code == 0
