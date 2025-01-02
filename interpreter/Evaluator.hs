module Evaluator where

import Control.Applicative
import Data.Bool

import Parser


{- +===============================+ -}
{- | Interpreter/Evaluator section | -}
{- +===============================+ -}


-- 
-- NOTE: in order to prevent collision, evaluation will take the first word
--       and decide which method to run in order to prevent Parsing error
-- 
getFirstWord :: String -> String
getFirstWord (x:xs) = if alfabet x then x : getFirstWord xs else []

{- Evaluation function  -}
evaluate = \input -> do
  if (getFirstWord input) == "tampilkan" then
    parse (tampilkan) input
  else if (getFirstWord input) == "jika" then
    parse (kondisi) input
  else
    parse (aritmatika) input


--
-- BNF for aritmatika (arithmetic)
--
-- <aritmatika> ::= <lpred>
-- <lpred>      ::= <hpred> + <lpred> | <hpred> - <lpred> | <hpred>
-- <hpred>      ::= <factor> * <hpred> | <factor> / <hpred> | <factor>
-- <factor>     ::= '(' <lpred> ')' | <bilanganAsli>

-- Lower precedence
lpred :: Parser Int
lpred = do
  n1 <- hpred
  satisfy (=='+')
  n2 <- lpred
  return (n1 + n2)
  <|> do
  -- Case for spaces in-between
  n1 <- hpred
  spasi
  satisfy (=='+')
  spasi
  n2 <- lpred
  return (n1 + n2)
  <|> do
  n1 <- hpred
  satisfy (=='-')
  n2 <- lpred
  return (n1 - n2)
  <|> do
  -- Case for spaces in-between
  n1 <- hpred
  spasi
  satisfy (=='-')
  spasi
  n2 <- lpred
  return (n1 - n2)
  <|> hpred

-- Higher precedence
hpred :: Parser Int
hpred = do
  n1 <- factor
  satisfy (=='*')
  n2 <- hpred
  return (n1 * n2)
  <|> do 
  n1 <- factor
  spasi
  satisfy (=='*')
  spasi
  n2 <- hpred
  return (n1 * n2)
  <|> do
  n1 <- factor
  satisfy (=='/')
  n2 <- hpred
  return (n1 `div` n2)
  <|> do
  n1 <- factor
  spasi
  satisfy (=='/')
  spasi
  n2 <- hpred
  return (n1 `div` n2)
  <|> factor

factor :: Parser Int
factor = do
  satisfy (=='(')
  result <- lpred
  satisfy (==')')
  return result
  <|> bilanganAsli

aritmatika :: Parser String
aritmatika = do
  result <- lpred
  return (show result)


--
-- BNF for boolean operation
-- 
-- <boolean> = <aritmatika> '<' <aritmatika> | <aritmatika> '>' <aritmatika> | <aritmatika> '=' <aritmatika>
--
applyBoolean :: Parser Bool
applyBoolean = do
  n1 <- aritmatika
  satisfy (=='<')
  n2 <- aritmatika
  return (n1 < n2)
  <|> do
  n1 <- aritmatika
  satisfy (=='>')
  n2 <- aritmatika
  return (n1 > n2)
  <|> do
  n1 <- aritmatika
  satisfy (=='=')
  n2 <- aritmatika
  return (n1 == n2)
  -- Cases for spaces in-between
  <|> do
  n1 <- aritmatika
  spasi
  satisfy (=='<')
  spasi
  n2 <- aritmatika
  return (n1 < n2)
  <|> do
  n1 <- aritmatika
  spasi
  satisfy (=='>')
  spasi
  n2 <- aritmatika
  return (n1 > n2)
  <|> do
  n1 <- aritmatika
  spasi
  satisfy (=='=')
  spasi
  n2 <- aritmatika
  return (n1 == n2)
  <|> return False

boolean :: Parser Bool
boolean = do
  result <- applyBoolean
  return result


--
-- BNF for printing statement
-- 
-- tampilkan ::= "tampilkan" <untaian> 
--
tampilkan :: Parser String
tampilkan = do
  perintah <- kataKunci  
  spasi
  s <- untaian
  if perintah == "tampilkan" then return s else return ""
  <|> return "ERROR: error occured in `tampilkan` statement"


--
-- BNF for if else statement
--
-- kondisi ::= "jika" <boolean> "maka" <tampilkan>
--
kondisi :: Parser String
kondisi = do
  katakunci1 <- kataKunci
  spasi
  bool <- boolean
  spasi
  katakunci2 <- kataKunci
  spasi
  result <- tampilkan
  if (katakunci1 == "jika") && (bool == True) && (katakunci2 == "maka") then return result else return ""
  <|> return "ERROR: error occured in `jika` statement"
