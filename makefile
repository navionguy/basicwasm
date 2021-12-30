./basicwasm : main.go \
		./filelist/filelist.go \
        ./fileserv/fileserv.go \
		./webmodules/gwbasic.wasm \
		./assets/js/wasm_exec.js \
		./assets/css/xterm.css \
		./assets/js/xterm.js \
		./assets/js/xterm.js.map \
		./assets/js/xterm-addon-fit.js \
		./assets/js/xterm-addon-fit.js.map
	go build -o basicwasm

./webmodules/gwbasic.wasm : ./webmodules/src/gwbasic/gwbasic.go \
			./ast/ast.go \
			./ast/program.go \
			./berrors/berrors.go \
			./builtins/builtins.go \
			./cli/cli.go \
			./decimal/decimal.go \
			./evaluator/evaluator.go \
			./evaluator/expressions.go \
			./filelist/filelist.go \
			./fileserv/fileserv.go \
			./gwtoken/gwtoken.go \
			./keybuffer/keybuffer.go \
			./lexer/lexer.go \
			./makefile \
			./object/object.go \
			./object/environ.go \
 			./parser/parser.go \
			./parser/parser_trace.go \
			./parser/parser_utils.go \
			./settings/settings.go \
			./token/token.go \
			./terminal/terminal.go \
#	tinygo build -no-debug -o ./webmodules/gwbasic.wasm -target=wasm ./webmodules/src/gwbasic/gwbasic.go
	GOOS=js GOARCH=wasm go build -ldflags "-s -w" -o ./webmodules/gwbasic.wasm ./webmodules/src/gwbasic/gwbasic.go

./assets/js/wasm_exec.js : /usr/local/go/misc/wasm/wasm_exec.js
	cp /usr/local/go/misc/wasm/wasm_exec.js ./assets/js/wasm_exec.js

#./assets/js/wasm_exec.js : /usr/local/go/misc/wasm/wasm_exec.js
#	cp /usr/local/go/misc/wasm/wasm_exec.js ./assets/js/wasm_exec.js

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
