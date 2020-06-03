./basicwasm : main.go \
        fileserv.go \
		./webmodules/gwbasic.wasm \
		./assets/js/wasm_exec.js \
		./assets/css/xterm.css \
		./assets/js/xterm.js \
		./assets/js/xterm.js.map \
		./assets/js/xterm-addon-fit.js \
		./assets/js/xterm-addon-fit.js.map
	go build -o basicwasm

./webmodules/gwbasic.wasm : ./webmodules/src/gwbasic/gwbasic.go
	GOOS=js GOARCH=wasm go build -o ./webmodules/gwbasic.wasm ./webmodules/src/gwbasic/gwbasic.go

./assets/js/wasm_exec.js : /usr/local/go/misc/wasm/wasm_exec.js
	cp /usr/local/go/misc/wasm/wasm_exec.js ./assets/js/wasm_exec.js

./assets/css/xterm.css : ~/node_modules/xterm/css/xterm.css
	cp ~/node_modules/xterm/css/xterm.css ./assets/css/xterm.css

./assets/js/xterm.js : ~/node_modules/xterm/lib/xterm.js
	cp ~/node_modules/xterm/lib/xterm.js ./assets/js/xterm.js

./assets/js/xterm.js.map : ~/node_modules/xterm/lib/xterm.js.map
	cp ~/node_modules/xterm/lib/xterm.js.map ./assets/js/xterm.js.map

./assets/js/xterm-addon-fit.js : ~/node_modules/xterm-addon-fit/lib/xterm-addon-fit.js
	cp ~/node_modules/xterm-addon-fit/lib/xterm-addon-fit.js ./assets/js/xterm-addon-fit.js

./assets/js/xterm-addon-fit.js.map : ~/node_modules/xterm-addon-fit/lib/xterm-addon-fit.js.map
	cp ~/node_modules/xterm-addon-fit/lib/xterm-addon-fit.js.map ./assets/js/xterm-addon-fit.js.map
