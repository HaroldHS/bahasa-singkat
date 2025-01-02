import Evaluator
import Parser


main :: IO()
main = do
  putStr "\n[*] File name: "
  fileName <- getLine
  fileContent <- readFile fileName
  let results = map (\input -> getParserResult $ evaluate input) (lines fileContent)
  mapM_ (\input -> if input /= "" then putStrLn input else putStr "") results
