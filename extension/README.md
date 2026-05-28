# PR Governor — Azure DevOps Extension

React-based Azure DevOps extension for managing PR Governor policies, viewing run history, and triggering simulations/executions.

## Prerequisites

- Node.js 20+
- A PocketBase backend instance running and accessible

## Setup

```bash
cd extension
npm install
```

## Development

```bash
# Start local dev server (port 3000)
npm run dev
```

To test in Azure DevOps, set `baseUri` in `vss-extension.json` to `"http://localhost:3000/"` and install the extension locally.

## Build

```bash
# Production build
npm run build

# Build + package as VSIX
npm run package
```

## Publish

1. Create a publisher at https://marketplace.visualstudio.com/manage/createpublisher
2. Update `publisher` in `vss-extension.json` with your publisher ID
3. Run `npm run package` to produce a `.vsix` file
4. Upload at https://marketplace.visualstudio.com/manage/publishers
5. Share with your Azure DevOps organization

## Extension Structure

```
extension/
├── vss-extension.json      # Extension manifest
├── src/
│   ├── index.tsx            # SDK init + React mount
│   ├── hub/
│   │   ├── HubRoot.tsx      # Tab navigation (Policies | Runs)
│   │   ├── policies/        # Policy list, detail, and form views
│   │   └── runs/            # Run history and detail views
│   ├── services/
│   │   └── pocketbase.ts    # PocketBase REST API client
│   └── models/
│       └── types.ts         # TypeScript types matching Go models
└── static/
    └── hub.html             # HTML shell for the hub contribution
```

## Configuration

The extension expects a PocketBase backend URL. By default it connects to `http://localhost:8090`. To change this, call `initClient(url)` before rendering, or update the `DEFAULT_BASE_URL` in `src/services/pocketbase.ts`.
