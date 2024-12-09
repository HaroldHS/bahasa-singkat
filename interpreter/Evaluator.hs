module Evaluator where

import Control.Applicative

import Parser


{- +===============================+ -}
{- | Interpreter/Evaluator section | -}
{- +===============================+ -}


evaluate :: Parser String
evaluate = aritmatika <|> tampilkan <|> kondisi


--
-- BNF for aritmatika (arithmetic)
--
-- aritmatika ::= lpred
-- lpred      ::= hpred + lpred | hpred - lpred | hpred
-- hpred      ::= factor * hpred | factor / hpred | factor
-- factor     ::= (lpred) | bilanganAsli

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
-- boolean = aritmatika '<' aritmatika | aritmatika '>' aritmatika | aritmatika '=' aritmatika
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
  <|> return ""


--
-- BNF for if else statement
--
-- kondisi ::= "jika" boolean "maka" tampilkan
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
  <|> return ""
