module BytecodeGenerator where

import Control.Applicative
import Data.Bool

import Parser


{- +============================+ -}
{- | Bytecode Generator section | -}
{- +============================+ -}


--
-- NOTE: This bytecode generator is enhanced form of /interpreter/Evaluator.hs.
--       So, the similarity of the code could be significant.
--

getFirstWord :: String -> String
getFirstWord (x:xs) = if alfabet x then x : getFirstWord xs else []


generate = \input -> do
  if (getFirstWord input) == "tampilkan" then
    parse (tampilkan) input
  else if (getFirstWord input) == "diberikan" then
    parse (variabel) input
  else
    parse (aritmatika) input


lpred :: Parser [Bytecode]
lpred = do
  n1 <- hpred
  satisfy (=='+')
  n2 <- lpred
  return $ concat [[TAMBAH], n1, n2]
  <|> do
  n1 <- hpred
  spasi
  satisfy (=='+')
  spasi
  n2 <- lpred
  return $ concat [[TAMBAH], n1, n2]
  <|> do
  n1 <- hpred
  satisfy (=='-')
  n2 <- lpred
  return $ concat [[KURANG], n1, n2]
  <|> do
  n1 <- hpred
  spasi
  satisfy (=='-')
  spasi
  n2 <- lpred
  return $ concat [[KURANG], n1, n2]
  <|> hpred

hpred :: Parser [Bytecode]
hpred = do
  n1 <- factor
  satisfy (=='*')
  n2 <- hpred
  return $ concat [[KALI], n1, n2]
  <|> do
  n1 <- factor
  spasi
  satisfy (=='*')
  spasi
  n2 <- hpred
  return $ concat [[KALI], n1, n2]
  <|> do
  n1 <- factor
  satisfy (=='/')
  n2 <- hpred
  return $ concat [[BAGI], n1, n2]
  <|> do
  n1 <- factor
  spasi
  satisfy (=='/')
  spasi
  n2 <- hpred
  return $ concat [[BAGI], n1, n2]
  <|> factor

factor :: Parser [Bytecode]
factor = do
  satisfy (=='(')
  result <- bilanganAsli
  satisfy (==')')
  return [BILANGAN result]
  <|> do
  result <- bilanganAsli
  return [BILANGAN result]

aritmatika :: Parser [Bytecode]
aritmatika = do
  result <- lpred
  return result


{- Command functions below -}

tampilkan :: Parser [Bytecode]
tampilkan = do
  perintah <- kataKunci  
  spasi
  s <- untaian
  if perintah == "tampilkan" then return [TAMPILKAN s] else return [DO_NOTHING]
  <|> return [ERROR "ERROR: error occured in `tampilkan` statement"]


variabel :: Parser [Bytecode]
variabel = do
  katakunci1 <- kataKunci
  spasi
  namavariabel <- some (satisfy alfabet)
  spasi
  katakunci2 <- kataKunci
  spasi
  n <- bilanganAsli
  if (katakunci1 == "diberikan") && (katakunci2 == "adalah") then return [VARIABEL_BILANGAN namavariabel n] else return [DO_NOTHING]
  <|> do
  katakunci1 <- kataKunci
  spasi
  namavariabel <- some (satisfy alfabet)
  spasi
  katakunci2 <- kataKunci
  spasi
  s <- untaian
  if (katakunci1 == "diberikan") && (katakunci2 == "adalah") then return [VARIABEL_UNTAIAN namavariabel s] else return [DO_NOTHING]
  <|> return [ERROR "ERROR: error occured in `diberikan` statement"]
