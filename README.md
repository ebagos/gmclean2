# gmclean: ファイルのハッシュ値によって重複ファイルを削除するプログラム

1. プログラム起動時に、検証するディレクトリ群を指定するJSONファイルを指定する
    - 与えられなかった場合は、カレントディレクトリにある「config.json」を読み込む
    - さらに、「config.json」が存在しない場合は、エラーを出力して終了する
    - JSONファイルの形式は以下の通り
        ```
       {
           "dirs": [
               "dir1",
               "dir2",
               "dir3"
           ]
       }
       ```
2. 処理後、各ディレクトリには、「results.json」というそのディレクトリ内のファイル情報を記録したJSONファイルを作成する
    - このファイルは、次回の実行時に、ファイルサイズとファイルの更新日付が同一の場合は、ハッシュ値の計算を省くために使用される
    - 「results.json」のJSONの構造は以下の通り
        ```
       {
           "files": [
               {
                   "path": "path/to/file1",
                   "size": 12345,
                   "date": "2020-01-01T00:00:00+09:00",
                   "hash": "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
               },
               {
                   "path": "path/to/file2",
                   "size": 12345,
                   "date": "2020-01-01T00:00:00+09:00",
                   "hash": "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
               },
               ...
           ]
       }
       ```
3. ディレクトリ内で、ファイル達の情報を獲得し、「results.json」を更新する
   - 全てのディレクトリで「results.json」の更新が終了した後、各ディレクトリの「results.json」を比較し、重複ファイルを削除し、方削除した内容を「results.json」に記録する
   - なお、同一ファイルのうち最新のファイルのみ残すこととする
