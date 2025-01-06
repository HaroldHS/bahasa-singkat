#!/bin/sh

#
# Name        : build.sh
# License     : BSD 3-Clause "New" or "Revised"
# Description : Bash script to build binary files for JIT compiler
#

help () {
	echo "[USAGE] build.sh ('build' / 'clean')\n"
}

if [ $# -eq 1 ]; then
	if [ $1 = "help" ]; then
		help
		exit 0
	elif [ $1 = "build" ]; then
		# Create 'artefacts' folder and copy all require files
		mkdir ./artefacts
		cp ./compiler/BytecodeGenerator.hs ./artefacts/BytecodeGenerator.hs
		cp ./compiler/Parser.hs ./artefacts/Parser.hs
		cp ./compiler/Main.hs ./artefacts/Main.hs
		cd ./artefacts
		ghc Main.hs -o compiler
		cd ../
		# NOTE: For engine, run 'go build' first then move the executable file into /artefacts
		cd ./engine
		go build
		mv ./engine ../artefacts/engine
		cd ../
	elif [ $1 = "clean" ]; then
		if [ -d ./artefacts ]; then
			rm -r ./artefacts
		fi
	else
		echo "[-] Invalid command: $1"
		exit 1
	fi
else
	help
	exit 1
fi
