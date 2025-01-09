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
  else if (getFirstWord input) == "jika" then
    parse (kondisi) input
  else if (getFirstWord input) == "pengulangan" then
    parse (pengulangan) input
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
  return [PUSH result]
  <|> do
  result <- bilanganAsli
  return [PUSH result]
  <|> do
  katakunci1 <- kataKunci
  spasi
  katakunci2 <- kataKunci
  spasi
  namavariabel <- namaVariabel
  if (katakunci1 == "variabel") && (katakunci2 == "bilangan") then return [GET_VARIABEL_BILANGAN namavariabel] else return [DO_NOTHING]

aritmatika :: Parser [Bytecode]
aritmatika = do
  result <- lpred
  return result

applyBoolean :: Parser [Bytecode]
applyBoolean = do
  n1 <- aritmatika
  satisfy (=='<')
  n2 <- aritmatika
  return $ concat [[LEBIH_KECIL], n1, n2]
  <|> do
  n1 <- aritmatika
  satisfy (=='>')
  n2 <- aritmatika
  return $ concat [[LEBIH_BESAR], n1, n2]
  <|> do
  n1 <- aritmatika
  satisfy (=='=')
  n2 <- aritmatika
  return $ concat [[SAMA_DENGAN], n1, n2]
  <|> do
  n1 <- aritmatika
  spasi
  satisfy (=='<')
  spasi
  n2 <- aritmatika
  return $ concat [[LEBIH_KECIL], n1, n2]
  <|> do
  n1 <- aritmatika
  spasi
  satisfy (=='>')
  spasi
  n2 <- aritmatika
  return $ concat [[LEBIH_BESAR], n1, n2]
  <|> do
  n1 <- aritmatika
  spasi
  satisfy (=='=')
  spasi
  n2 <- aritmatika
  return $ concat [[SAMA_DENGAN], n1, n2]
  <|> return [SALAH]

boolean :: Parser [Bytecode]
boolean = do
  result <- applyBoolean
  return result

{- Command functions below -}

tampilkan :: Parser [Bytecode]
tampilkan = do
  perintah <- kataKunci
  spasi
  s <- untaian
  if perintah == "tampilkan" then return [TAMPILKAN s] else return [DO_NOTHING]
  <|> do
  perintah <- kataKunci
  spasi
  result <- aritmatika
  if perintah == "tampilkan" then return $ concat [[TAMPILKAN_FROM_STACK, RETURN], result] else return [DO_NOTHING]
  <|> do
  perintah <- kataKunci  
  spasi
  katakunci1 <- kataKunci
  spasi
  -- TODO: Fix the error due to commented codes below
  {-
  katakunci2 <- kataKunci
  spasi
  -}
  namavariabel <- namaVariabel
  if (perintah == "tampilkan") && (katakunci1 == "variabel") {-&& (katakunci2 == "untaian")-} then return [TAMPILKAN_FROM_STACK, RETURN, GET_VARIABEL_UNTAIAN namavariabel] else return [DO_NOTHING]
  <|> return [ERROR "ERROR: error occured in `tampilkan` statement"]


variabel :: Parser [Bytecode]
variabel = do
  katakunci1 <- kataKunci
  spasi
  namavariabel <- some (satisfy alfabet)
  spasi
  katakunci2 <- kataKunci
  spasi
  s <- untaian
  if (katakunci1 == "diberikan") && (katakunci2 == "adalah") then return [SET_VARIABEL_UNTAIAN namavariabel s] else return [DO_NOTHING]
  <|> do
  katakunci1 <- kataKunci
  spasi
  namavariabel <- namaVariabel
  spasi
  katakunci2 <- kataKunci
  spasi
  n <- bilanganAsli
  if (katakunci1 == "diberikan") && (katakunci2 == "adalah") then return [SET_VARIABEL_BILANGAN namavariabel n] else return [DO_NOTHING]
  <|> return [ERROR "ERROR: error occured in `diberikan` statement"]


kondisi :: Parser [Bytecode]
kondisi = do
  katakunci1 <- kataKunci
  spasi
  bool <- boolean
  spasi
  katakunci2 <- kataKunci
  spasi
  result <- tampilkan
  if (katakunci1 == "jika") && (katakunci2 == "maka") then return $ concat [[END_JIKA_BLOCK], result, bool] else return [DO_NOTHING]
  <|> return [ERROR "ERROR: error occured in `jika` statement"]

pengulangan :: Parser [Bytecode]
pengulangan = do
  katakunci1 <- kataKunci
  spasi
  katakunci2 <- kataKunci
  spasi
  jumlah <- aritmatika
  spasi
  katakunci3 <- kataKunci
  spasi
  result <- tampilkan
  if (katakunci1 == "pengulangan") then return $ concat [[END_PENGULANGAN_BLOCK], result, [PENGULANGAN], jumlah] else return [DO_NOTHING]
  <|> return [ERROR "ERROR: error occured in `pengulangan` statement"]

