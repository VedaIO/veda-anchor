// TypeScript declarations for Wails runtime
declare global {
  interface Window {
    runtime?: {
      BrowserOpenURL(url: string): void;
    };
    go: {
      main: {
        App: {
          AddWebBlocklist(arg1: string): Promise<void>;
          BlockApps(arg1: Array<string>): Promise<void>;
          CheckChromeExtension(): Promise<boolean>;
          OpenBrowser(url: string): Promise<void>;
          ClearAppBlocklist(): Promise<void>;
          ClearWebBlocklist(): Promise<void>;
          DisableAutostart(): Promise<void>;
          EnableAutostart(): Promise<void>;
          GetAppBlocklist(): Promise<Array<any>>;
          GetAppDetails(arg1: string): Promise<any>;
          GetAppLeaderboard(arg1: string, arg2: string): Promise<Array<any>>;
          GetAutostartStatus(): Promise<boolean>;
          GetIsAuthenticated(): Promise<boolean>;
          GetWebBlocklist(): Promise<Array<any>>;
          GetWebDetails(arg1: string): Promise<any>;
          GetWebLeaderboard(arg1: string, arg2: string): Promise<Array<any>>;
          GetWebLogs(
            arg1: string,
            arg2: string,
            arg3: string
          ): Promise<Array<any>>;
          HasPassword(): Promise<boolean>;
          LoadAppBlocklist(arg1: Array<number>): Promise<void>;
          LoadWebBlocklist(arg1: Array<number>): Promise<void>;
          Login(arg1: string): Promise<boolean>;
          Logout(): Promise<void>;
          RegisterExtension(arg1: string): Promise<void>;
          RemoveWebBlocklist(arg1: string): Promise<void>;
          SaveAppBlocklist(): Promise<Array<number>>;
          SaveWebBlocklist(): Promise<Array<number>>;
          Search(arg1: string, arg2: string, arg3: string): Promise<Array<any>>;
          SetPassword(arg1: string): Promise<void>;
          Stop(): Promise<void>;
          UnblockApps(arg1: Array<string>): Promise<void>;
          Uninstall(arg1: string): Promise<void>;
        };
      };
    };
  }
}

export {};
