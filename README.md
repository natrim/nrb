## NRB

- just simple build script for react app's

### Installation
```shell
go install github.com/natrim/nrb/cmd/nrb@latest
```

##### Note:
index.html must contain index.js/css import for watch mode to work
ie.
```
<link rel="stylesheet" href="/assets/index.css">
<script type="module" src="/assets/index.js"></script>
```

### TODO

- autoinject main js/css in watch mode
- configuration (toml config?) 
- custom stuff
- prebuilt bin (npm compat)
