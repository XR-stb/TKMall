#!/usr/bin/env python3

import argparse

# import argcomplete

import sys
import os
import time

import scripts.python.log

# 导入其他任意路径下的py模块
project_root = os.path.dirname(os.path.abspath(__file__))
sys.path.append(project_root + "/scripts/python")
try:
    import scripts.python.util as util
    from scripts.python.log import logger
except:
    pass

def get_golang_servers():
    return util.get_directories("./cmd/")


class Make:
    def __init__(self, args):
        self.args = args
        self.golang_targets = get_golang_servers()
        self.all_targets = self.golang_targets

    def run(self):
        # 反射到成员函数上进行调用，注意起名规范
        action = "_" + self.args.action + "_"
        getattr(self, action)()

    def build_targets(self, targets):
        targets_success = []
        targets_failed = []
        for target in targets:
            build_ret = 114514
            if target in self.golang_targets:
                build_ret = util.exec_cmd_with_color(
                    "go build -o build/bin/ ./cmd/%s"
                    % target
                )

            if build_ret == 114514:
                raise Exception("unkown tartget: %s" % target)

            if build_ret != 0:
                targets_failed.append(str(target))
            else:
                targets_success.append(str(target))

        if targets_success:
            logger.Info("Success targets: " + ", ".join(targets_success))
        if targets_failed:
            logger.Error("Failed targets: " + ", ".join(targets_failed))

    def _build_(self):
        targets = self.args.targets
        if targets == "all" or targets == ["all"]:
            self.build_targets(self.all_targets)
        else:
            self.build_targets(targets)

    def _run_(self):
        # 先启动 etcd, 避免服务注册时 阻塞住
        logger.Info("Starting etcd...")
        util.exec_cmd_with_color("bash ./tools/etcd/start.sh")

        targets = self.args.targets
        if targets == "all" or targets == ["all"]:
            self._kill_()

            # 优先运行gateway
            if "gateway" in self.golang_targets:
                logger.Info("Starting gateway first in background...")
                util.exec_cmd_with_color("./build/bin/gateway &")
                util.exec_cmd_with_color("echo \"\n\"")
                time.sleep(2)
                # 运行其他服务（排除gateway）
                other_services = [s for s in self.golang_targets if s != "gateway"]
                for service in other_services:
                    logger.Info(f"Starting {service} in background...")
                    util.exec_cmd_with_color(f"./build/bin/{service} &")
                return
            # 如果不存在gateway则按原顺序运行
            for service in self.golang_targets:
                util.exec_cmd_with_color(f"./build/bin/{service}")
        else:
            # 单个服务运行时也需要确保 etcd 启动
            logger.Info("Starting etcd...")
            util.exec_cmd_with_color("bash ./tools/etcd/start.sh")
            
            for target in targets:
                if target in self.golang_targets:
                    util.exec_cmd_with_color("./build/bin/%s" % target)
                    return
                raise Exception("unknown targets %s", target)
    
    def _install_(self):
        self._build_()
        self._run_()

    def _genproto_(self):
        util.exec_cmd_with_color("bash ./scripts/shell/gen_proto.sh")

    def _unitest_(self):
        """运行单元测试"""
        util.exec_cmd_with_color("bash ./scripts/shell/run_tests.sh")

    def _kill_(self):
        util.exec_cmd_with_color("bash ./scripts/shell/kill_servers.sh")


def parse_args():
    # https://docs.python.org/zh-cn/3.6/library/argparse.html#argparse.ArgumentParser.add_argument
    parser = argparse.ArgumentParser(
        description="Unified project build script", prog="TKMall"
    )
    # parser.add_argument('integers', metavar='1', type=int, nargs='+', help='an integer for the accumulator')
    # parser.add_argument('--sum', dest='accumulate', action='store_const', const=sum, default=max, help='sum the integers (default: find the max)')
    # parser.print_help()

    subparsers = parser.add_subparsers(dest="action", help="subcommand help")

    build_parser = subparsers.add_parser("build", help="build a target")
    build_parser.add_argument(
        "targets",
        help="the server name, default is all",
        default="all",
        metavar="gateway",
        type=str,
        nargs="*",
        choices=["all"] + get_golang_servers(),
    )

    run_parser = subparsers.add_parser("run", help="run a target")
    run_parser.add_argument(
        "targets",
        help="the server name, default is all",
        default="all",
        metavar="gateway",
        type=str,
        nargs="*",
        choices=["all"] + get_golang_servers(),
    )

    install_parser = subparsers.add_parser("install", help="build && run a target")
    install_parser.add_argument(
        "targets",
        help="the server name, default is all",
        default="all",
        metavar="gateway",
        type=str,
        nargs="*",
        choices=["all"] + get_golang_servers(),
    )

    proto_parser = subparsers.add_parser("genproto", help="gen proto")

    kill_parser = subparsers.add_parser("kill", help="kill all running servers")
    
    unittest_parser = subparsers.add_parser("unitest", help="run unit tests for all services")

    # https://kislyuk.github.io/argcomplete/
    # TODO: 命令行自动补全命令, 这个autocomplete包不好下，回头弄
    # argcomplete.autocomplete(parser)
    return parser.parse_args()


def main():
    args = parse_args()
    Make(args).run()


if __name__ == "__main__":
    main()
