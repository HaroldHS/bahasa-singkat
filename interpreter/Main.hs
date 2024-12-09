import Evaluator
import Parser


main :: IO()
main = do
  putStr "\n[*] File name: "
  fileName <- getLine
  fileContent <- readFile fileName
  let results = map (\input -> getParserResult $ parse (evaluate) input) (lines fileContent)
  mapM_ putStrLn results

