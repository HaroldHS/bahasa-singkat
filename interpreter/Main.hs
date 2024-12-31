import Evaluator
import Parser


main :: IO()
main = do
  putStr "\n[*] File name: "
  fileName <- getLine
  fileContent <- readFile fileName
  let diberikan_strings = filter (\input -> checkParserTypeForFilter $ parse (variabel) input) (lines fileContent)
  let vars = map (parse (variabel)) diberikan_strings
  let results = map (\input -> getParserResult $ evaluate input) (lines fileContent)
  mapM_ (\input -> if input /= "" then putStrLn input else putStr "") results
