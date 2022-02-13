db.createCollection("test_db", { size: 104857600 } )
doc={'a':0, 'item':9}
db.test_db.save(doc)
doc={'a':1, 'item':99}
db.test_db.save(doc)
db.test_db.find();
