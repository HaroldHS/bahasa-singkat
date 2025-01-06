module Parser where

import Control.Applicative
import Data.Bool

{- Datatype for bytecode generation -}
data Bytecode = RETURN | DO_NOTHING
                       | PUSH Int
                       | BILANGAN Int
                       | UNTAIAN String
                       | SET_VARIABEL_BILANGAN String Int
                       | SET_VARIABEL_UNTAIAN String String
                       | GET_VARIABEL_BILANGAN String
                       | GET_VARIABEL_UNTAIAN String
                       | TAMPILKAN String
                       | TAMPILKAN_FROM_STACK
                       | TAMBAH
                       | KURANG
                       | KALI
                       | BAGI
                       | LEBIH_KECIL
                       | LEBIH_BESAR
                       | SAMA_DENGAN
                       | BENAR
                       | SALAH
                       | END_IF
                       | ERROR String
                       deriving (Show, Eq)

{- +================+ -}
{- | Parser section | -}
{- +================+ -}

newtype Parser a = P (String -> Maybe (a, String))

parse :: Parser a -> String -> Maybe (a, String)
parse (P p) input = p input

{- Instances for parser  -}
instance Functor Parser where
  fmap f (P p) = P $ \input -> do
    (result, rest) <- p input
    Just(f result, rest)

instance Applicative Parser where
  pure a = P $ \input -> Just(a, input)
  (P f) <*> (P a) = P $ \input -> do
    (f, resti1) <- f input
    (i, resti2) <- a resti1
    Just(f i, resti2)

instance Alternative Parser where
  empty = P $ \_ -> Nothing
  (P p1) <|> (P p2) = P $ \input -> p1 input <|> p2 input

instance Monad Parser where
  p >>= f = P $ \input -> case parse p input of Nothing -> Nothing
                                                Just(a, rest) -> parse (f a) rest

{- Auxiliary functions for parser -}
getc :: Parser Char
getc = P $ \input -> case input of []     -> Nothing
                                   (x:xs) -> Just (x, xs)

satisfy :: (Char -> Bool) -> Parser Char
satisfy f = do
  c <- getc
  if f c then return c else empty

getParserResult :: Maybe([Bytecode], String) -> [Bytecode]
getParserResult Nothing = []
getParserResult (Just (r, ri)) = r

{- Helper functions for parser -}
angka c | c `elem` ['0' .. '9'] = True
        | otherwise             = False

operator c | c `elem` ['+', '-', '*', '/'] = True
           | otherwise                     = False

alfabet c | c `elem` ['a' .. 'z'] = True
          | c `elem` ['A' .. 'Z'] = True
          | otherwise             = False

-- Whitelisted characters for untaian (string)
karakter c | c `elem` ['0' .. '9']                   = True
           | c `elem` ['a' .. 'z']                   = True
           | c `elem` ['A' .. 'Z']                   = True
           | c `elem` [' ', ',', '.', '?', '!', '"'] = True
           | otherwise                               = False

-- Preserved keyword
kata s | s == "tampilkan" = True
       | s == "diberikan" = True
       | s == "variabel"  = True
       | s == "adalah"    = True
       | s == "jika"      = True
       | s == "maka"      = True
       | otherwise        = False

-- Allowed characters for variable name
karakterVariabel c | c `elem` ['a' .. 'z'] = True
                   | c `elem` ['A' .. 'Z'] = True
                   | c `elem` ['0' .. '9'] = True
                   | c == '_'              = True
                   | otherwise             = False

{- Parser methods -}
spasi :: Parser String
spasi = do
  some (satisfy (==' '))
  return ""

operasi :: Parser Char
operasi = do
  o <- satisfy operator
  return o

bilanganAsli :: Parser Int
bilanganAsli = do
  n <- some (satisfy angka)
  return (read n :: Int)

bilanganBulat :: Parser Int
bilanganBulat = do
  satisfy (=='-')
  n <- bilanganAsli
  return (-n)
  <|> bilanganAsli

untaian :: Parser String
untaian = do
  satisfy (=='\'')
  u <- some (satisfy karakter)
  satisfy (=='\'')
  return (u)

kataKunci :: Parser String
kataKunci = do
  target <- some (satisfy alfabet)
  if kata target then return (target) else return ""

namaVariabel :: Parser String
namaVariabel = do
  result <- some (satisfy karakterVariabel)
  return result
