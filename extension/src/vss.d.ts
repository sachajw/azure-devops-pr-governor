/// <reference types="react" />
/// <reference types="react-dom" />

declare module "azure-devops-extension-sdk" {
  export interface IExtensionInitOptions {
    loaded?: boolean;
    applyTheme?: boolean;
  }

  export function init(options?: IExtensionInitOptions): Promise<void>;
  function ready(): Promise<void>;
  function notifyLoadSucceeded(): Promise<void>;
  function notifyLoadFailed(error: Error): Promise<void>;
  function getHost(): IHostContext;
  function getExtensionContext(): IExtensionContext;
  function getWebContext(): IWebContext;
  function getContributionId(): string;
  function getAccessToken(): Promise<string>;
  function resize(width?: number, height?: number);
}
