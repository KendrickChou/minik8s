import sys
from importlib import import_module


def main():
    args = sys.argv[1:]
    func_name = args[0]
    func_args = args[1:]

    module = import_module("function")

    function = getattr(module, func_name)
    res = function(*func_args)

    if res is None or isinstance(res, int) or isinstance(res, bool) or isinstance(res, str):
        print(1, res)
    else:
        print(0, "Unspported Return Type")


if __name__ == '__main__':
    main()
