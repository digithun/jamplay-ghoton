# [jamplay-adore] [v1.0.2]

This repository purpose is to collect hearts

## API

## GET /asset/`:AssetType`/`:ContentType`/`:ContentID`
#### Header:
```
Nothing..
```

#### Params:
```
AssetType : String // eg. heart, banana, coffee
ContentType : String // eg. episode, book, author
ContentId : String // eg. bookId, episodeId, authorId
```
#### Result:
```
{
    "result": {
        "assetType": "heart"
        "contentType": "episode",
        "contentId": "2sdf28wer79520088d479",
        "count": 45642
    }
}
```

## PUT /asset/:`AssetType`/:`ContentType`/:`ContentID`/:`Amount`
#### Header:
```
UserID: String // eg your user id d9f8b990rf8800d0v990
```

#### Params:
```
AssetType : String // eg. heart, banana, coffee
ContentType : String // eg. episode, book, author
ContentID : String // eg. bookId, episodeId, authorId
Amount: Integer // eg 5, will send 5 times, default is 1 if not send
```

#### Result:
```
{
    "result": {
        "assetType": "heart"
        "contentType": "episode",
        "contentId": "2sdf28wer79520088d479",
        "count": 45647 // will increment
    }
}
```

