from colorama import init, Fore, Style
import threading

class Logger:
    _instance = None
    _lock = threading.Lock()

    def __new__(cls, *args, **kwargs):
        if not cls._instance:
            with cls._lock:
                if not cls._instance:
                    cls._instance = super(Logger, cls).__new__(cls, *args, **kwargs)
                    cls._instance._initialized = False
        return cls._instance

    def __init__(self):
        if not self._initialized:
            # Initialize colorama
            init(autoreset=True)
            self._initialized = True

    def Info(self, message):
        print(Fore.GREEN + message)

    def Warn(self, message):
        print(Fore.YELLOW  + message)

    def Error(self, message):
        print(Fore.RED + message)

    def Debug(self, message):
        print(Fore.CYAN + message)

logger = Logger()
