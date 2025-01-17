import BytecodeGenerator
import Parser

import System.Environment (getArgs)

bytecodeListToFile fn ls = do
  mapM_ (\bytecode -> appendFile fn (show bytecode ++ "\n")) ls
  -- append an empty line to seperate the bytecodes of each line
  appendFile fn "\n"

main :: IO()
main = do
  args        <- getArgs
  fileContent <- readFile (head args)

  let results = map (\input -> getParserResult $ generate input) (filter (/= "") (lines fileContent))
  mapM_ (\each_bytecode -> if each_bytecode /= [] then bytecodeListToFile (head (tail args)) each_bytecode else appendFile (head (tail args)) "") results

  return ()
