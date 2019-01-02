# Faces in the macOS photos library

My photo collection is a mess. Getting it  under control has proven a challenge in terms of de-dupe, metadata cleanup and just general organization. I've tried numerous photo management apps and online backup strategies over the years. That's resulted in many different backups scattered across various thumbdrives and SD cards.

One such backup is a snapshot of the `*.photoslibrary` folder from my MacBook. This is where the builtin Photos.app (formerly iPhoto) stores files and metadata. I mostly used it for syncing photos off my phone, not as an organizing tool. Nevertheless, I wanted to explore what I could extract from the archive beyond image files. In particular, any facial recognition data that was automatically generated since that was a feature of the app.

The rest of this post details cracking open the internal databases and piecing together the data model. It uses standard unix/shell tools and assumes knowledge of SQL. I've also started a go package for inspecting and manipulating the photos library. If you're only interested in the latter, take a look at the *experimental* [`phace` project][phace] on github. Enough rambling, now for some technical details.

<p class="caution">CAUTION: Make a backup copy of your photo library before following along and use that. This avoids the possibility of any accidental data loss or corruption. You've been warned.</p>

Lets start by looking at the contents of the photo library. By default it's at `~/Pictures/Photos Library.photoslibrary`.

```
$ cd path/to/photos.photoslibrary
$ ls
Attachments/
Masks/
Masters/
Plugins/
Previews/
Thumbnails/
database/
private/
resources/
```

The `Masters` directory is where all the originals are kept, organized by date.

<a name="tree-masters"></a>
```
$ tree Masters | head -15
Masters
`-- 2016
    |-- 05
    |   `-- 07
    |       `-- 20160507-030928
    |           |-- IMG_0031.JPG
    |           `-- IMG_0032.JPG
    `-- 11
        |-- 01
        |   `-- 20161101-021106
        |       |-- Clip0.mov
        |       |-- Clip1.mov
        |       |-- Clip2.mov
        |       |-- DSC00001.JPG
        |       |-- DSC00003.JPG
```

What's more interesting for this post is the `database` directory.

```
$ cd database
$ ls
DataModelVersion.plist
ImageProxies.apdb
ImageProxies.apdb-wal
Library.apdb
Person.db
Properties.apdb
RKAlbum_name.skindex
RKVersion_searchIndexText.skindex
metaSchema.db
```

Based on past experience, I'm sure these are sqlite database files given the `.db` extensions and presence of WAL (Write-Ahead-Log) files. The `.apdb` is likely a custom extension someone at Apple made up. This can be validated with the `file` utility, the tool of choice for interrogated unknown file types.

```
$ file *
DataModelVersion.plist:            XML 1.0 document, ASCII text
ImageProxies.apdb:                 SQLite 3.x database, last written using SQLite version 3008008
ImageProxies.apdb-wal:             empty
Library.apdb:                      SQLite 3.x database, last written using SQLite version 3025002
Person.db:                         SQLite 3.x database, last written using SQLite version 3025002
Properties.apdb:                   SQLite 3.x database, last written using SQLite version 3008008
RKAlbum_name.skindex:              data
RKVersion_searchIndexText.skindex: data
metaSchema.db:                     SQLite 3.x database, last written using SQLite version 3008008
```

It's interesting to note that `Library.apdb` and `Person.db` were last modified by newer versions of sqlite. That's relavent as these turn out to be the two databases we're interested in for the bulk of this post. However, before getting ahead of ourselves we should look at the structure of these databases to guide us. I'll stick with the standard and near ubiquitious `sqlite3` CLI tool.

Lets start by looking at the tables in each of the databases. One approach is to load up the interactive shell and execute the appropriately named `.tables` command.

```
$ sqlite3
SQLite version 3.25.1 2018-09-18 20:20:44
Enter ".help" for usage hints.
Connected to a transient in-memory database.
Use ".open FILENAME" to reopen on a persistent database.
sqlite> .open Properties.apdb
sqlite> .tables
Array_VirtualReader          RKPlace_RTree_rowid
RKPlace                      RKPlace_VirtualBufferReader
RKPlace_RTree                RKPlace_modelId_RidIndex
RKPlace_RTree_node           RidList_VirtualReader
RKPlace_RTree_parent
```

Not much interesting in this first database. For the rest we'll use a shell loop over each `*db` named file since `sqlite3` accepts a database and SQL commands as arguments. The trailing `grep` is a hack to filter out a large number virtual tables and index tables of similar names that otherwise clutter the output.

```
$ for f in *db; do; echo "=== $f ==="; sqlite3 $f .tables; echo; done | grep -v '^RK.*_'
=== ImageProxies.apdb ===
Array_VirtualReader
RKCloudResource
RKImageProxyState
RKModelResource
RidList_VirtualReader

=== Library.apdb ===
Array_VirtualReader
RKAdjustmentData
RKAdminData
RKAlbum
RKAlbumVersion
RKAttachment
RKBookmark
RKCustomSortOrder
RKFolder
RKImageMask
RKImportGroup
RKKeyword
RKKeywordForVersion
RKMaster
RKMoment
RKMomentCollection
RKMomentYear
RKPlaceForVersion
RKVersion
RKVolume
RidList_VirtualReader

=== Person.db ===
Array_VirtualReader
RKFace
RKFaceGroup
RKFaceGroupFace
RKFacePrint
RKPerson
RKPersonVersion
RidList_VirtualReader

=== Properties.apdb ===
Array_VirtualReader          RKPlace_RTree_rowid

=== metaSchema.db ===
Array_VirtualReader    LiPropertyDef          LiTableDef
LiGlobals              LiPropertyHistory      RidList_VirtualReader
LiLibHistory           LiStringAtom
```

The table names here seem self-descriptive. In particular, the `RKFace` table in `Person.db` looks interesting. Lets take a look at that with the `.schema` command. Again ignoring some output with `grep -v` but also using `tr` to make it more legible. The latter replaces all the commas with newlines, converting the single line `CREATE TABLE` statement into a more readable list.

```
$ sqlite3 Person.db '.schema RKFace' | grep -v TRIGGER | tr , \\n
CREATE TABLE RKFace (modelId integer primary key autoincrement
 uuid varchar
 isInTrash integer
 personId integer
 cloudLibraryState integer
 hasBeenSynced integer
 adjustmentUuid varchar
 imageId varchar
 sourceWidth integer
 sourceHeight integer
 centerX decimal
 centerY decimal
 size decimal
 leftEyeX decimal
 leftEyeY decimal
 rightEyeX decimal
 rightEyeY decimal
 mouthX decimal
 mouthY decimal
 hidden integer
 manual integer
 hasSmile integer
 isBlurred integer
 isLeftEyeClosed integer
 isRightEyeClosed integer
 pose decimal
 masterIdentifier varchar
 masterSourceWidth integer
 masterSourceHeight integer
 masterCenterX decimal
 masterCenterY decimal
 masterSize decimal
 masterLeftEyeX decimal
 masterLeftEyeY decimal
 masterRightEyeX decimal
 masterRightEyeY decimal
 masterMouthX decimal
 masterMouthY decimal
 nameSource integer
 isHiddenInGroups integer);
CREATE INDEX RKFace_uuid_index on RKFace(uuid);
CREATE INDEX RKFace_personId_index on RKFace(personId);
CREATE INDEX RKFace_imageId_index on RKFace(imageId);
```

Bingo! Promising column names that look like facial coordinates such as `leftEyeX` and image sizes. What's missing is something resembling a bounding box for the face. But there is the ambigously named `size` column. I'm guessing that's the radius or diameter about `(centerX, centerY)` of a face. With these assumptions lets `SELECT` a record to take a closer look. The `-line` option is key to generating readable output with this many fields vs the default column layout.

```
$ sqlite3 -line Person.db 'select * from RKFace limit 1'
           modelId = 1
              uuid = WG3y7TXURKa2zPkNz7P3tg
         isInTrash = 0
          personId = 0
 cloudLibraryState = NULL
     hasBeenSynced = 0
    adjustmentUuid = UNADJUSTEDNONRAW
           imageId = 49RT+c2xRB6yWYs8attNuA
       sourceWidth = 6000
      sourceHeight = 4000
           centerX = 0.397666666666667
           centerY = 0.39475
              size = 0.108
          leftEyeX = 0.378166666666667
          leftEyeY = 0.4215
         rightEyeX = 0.4165
         rightEyeY = 0.41975
            mouthX = 0.394833333333333
            mouthY = 0.35175
            hidden = 0
            manual = 0
          hasSmile = 0
         isBlurred = 1
   isLeftEyeClosed = 0
  isRightEyeClosed = 0
              pose = -0.233882412314415
  masterIdentifier = NULL
 masterSourceWidth = 6000
masterSourceHeight = 4000
     masterCenterX = 0.397666666666667
     masterCenterY = 0.39475
        masterSize = 0.108
    masterLeftEyeX = 0.378166666666667
    masterLeftEyeY = 0.4215
   masterRightEyeX = 0.4165
   masterRightEyeY = 0.41975
      masterMouthX = 0.394833333333333
      masterMouthY = 0.35175
        nameSource = 0
  isHiddenInGroups = 0
```

All the X, Y coordinates are decimals in the range 0-1 implying a normalized coordinate space. That means one corner should be (0,0) and the opposite diagonal (1,1). Mapping to pixels is just a matter of scaling by the height and width.[^coords]

Next step is to find the actual image file associated with this face record. The `imageId` field looks like a SQL foriegn key to another table, especially since it's an indexed column based on the `.schema` output

```
CREATE INDEX RKFace_imageId_index on RKFace(imageId);
```

The question is, which table should be `JOIN`-ed against? Lets start by looking at an `RKMaster` record from `Library.apdb` given the name aligns with the `Masters/` directory.

```
$ sqlite3 -line Library.apdb 'SELECT * FROM RKMaster LIMIT 1'
                   modelId = 1
                      uuid = LnGAi0CuSr+vPGWunW26tA
               fingerprint = AVujFE+Yg51mLA6iKEcwIQ2n14XU
               orientation = 6
                      name = IMG_0031.JPG
                createDate = 484283367.792741
                 isInTrash = 0
               inTrashDate = NULL
         cloudLibraryState = 0
             hasBeenSynced = 0
        isCloudQuarantined = 0
            fileVolumeUuid = NULL
           fileIsReference = 0
                 isMissing = 0
                  duration = NULL
      fileModificationDate = 468983941
                bookmarkId = NULL
                  fileSize = 1906431
                     width = 4032
                    height = 3024
                       UTI = public.jpeg
           importGroupUuid = NP4olZ57Q1aA1vriXknJJg
       alternateMasterUuid = NULL
       originalVersionName = IMG_0031.JPG
                  fileName = IMG_0031.JPG
      isExternallyEditable = 0
                isTrulyRaw = 0
            hasAttachments = 0
                  hasNotes = 0
                 imagePath = 2016/05/07/20160507-030928/IMG_0031.JPG
                 imageDate = 468994741.348
          fileCreationDate = 468983941
          originalFileName = IMG_0031.JPG
          originalFileSize = 1906431
                importedBy = 3
                 burstUuid = NULL
            importComplete = 1
imageTimeZoneOffsetSeconds = -28800
          photoStreamTagId = NULL
              mediaGroupId = NULL
    hasCheckedMediaGroupId = 1
```

Sure enough, there's an `imagePath` field with a relative filename string. Furthermore, it matches one of the images in the `tree` output of the `Masters/` directory [above](#tree-masters). We're on the right track.

Now to use `RKFace.imageId` to find the matching `RKMaster` record. `RKMaster.uuid` is the same length with a similar set of random alpha-numeric characters.[^uuids]

```
$ sqlite3 Library.apdb "SELECT count(*) FROM RKMaster WHERE uuid = '49RT+c2xRB6yWYs8attNuA'"
count(*)
0
```

No records so `RKFace.imageId` isn't a foreign key to `RKMaster`. Given the small database and lack of another table to check we'll resort to a brute force search. That'll rely on the sqlite `.dump` command which outputs raw SQL commands that can reconstruct an exact clone of a given database. Piping the result through `grep` with our given id should find the record.

```
$ sqlite3 Library.apdb .dump | grep '49RT+c2xRB6yWYs8attNuA'
INSERT INTO RKVersion VALUES(6,'49RT+c2xRB6yWYs8attNuA',1,NULL,NULL,499659066.52122598075,0,0,NULL,0,-1,0,0,NULL,0,2,'UNADJUSTEDNONRAW','DSC00004.JPG',0,496452726,NULL,-25200,NULL,0,499659066.52122802356,1,'xoCatkSMRueQxFJSb1kxpw',6,NULL,'xoCatkSMRueQxFJSb1kxpw','GMT-0700',0,0,1,4000,6000,4000,6000,NULL,0,NULL,NULL,NULL,NULL,NULL,4,0,NULL,NULL,NULL,1,NULL,NULL,NULL,4,0,NULL,NULL,32,0,'H8M6+2XgSYaziE9rUY2WdQ',0,NULL,1,'UNADJUSTEDNONRAW',NULL,'UNADJUSTEDNONRAW',NULL,NULL,NULL,NULL,NULL,NULL,2,0);
$ sqlite3 Library.apdb '.schema RKVersion' | tr , \\n | head -3
CREATE TABLE RKVersion (modelId integer primary key autoincrement
 uuid varchar
 orientation integer
```

So, `RKVersion` is the table we're after and it's second column is also named `uuid`. Lets cleanup the output and determine how to join it with `RKMaster`.

```
$ sqlite3 -line Library.apdb "SELECT * FROM RKVersion WHERE uuid = '49RT+c2xRB6yWYs8attNuA'"
                        modelId = 6
                           uuid = 49RT+c2xRB6yWYs8attNuA
                    orientation = 1
                naturalDuration = NULL
                           name = NULL
                     createDate = 499659066.521226
                     isFavorite = 0
                      isInTrash = 0
                    inTrashDate = NULL
                       isHidden = 0
                colorLabelIndex = -1
              cloudLibraryState = 0
                  hasBeenSynced = 0
                cloudIdentifier = NULL
             isCloudQuarantined = 0
                           type = 2
                 adjustmentUuid = UNADJUSTEDNONRAW
                       fileName = DSC00004.JPG
                       hasNotes = 0
                      imageDate = 496452726
                      burstUuid = NULL
     imageTimeZoneOffsetSeconds = -25200
            reverseLocationData = NULL
     reverseLocationDataIsValid = 0
               lastModifiedDate = 499659066.521228
                  versionNumber = 1
                     masterUuid = xoCatkSMRueQxFJSb1kxpw
                       masterId = 6
                  rawMasterUuid = NULL
               nonRawMasterUuid = xoCatkSMRueQxFJSb1kxpw
              imageTimeZoneName = GMT-0700
                     mainRating = 0
                      isFlagged = 0
                     isEditable = 1
                   masterHeight = 4000
                    masterWidth = 6000
                processedHeight = 4000
                 processedWidth = 6000
                       rotation = NULL
                 hasAdjustments = 0
                overridePlaceId = NULL
                       latitude = NULL
                      longitude = NULL
                   exifLatitude = NULL
                  exifLongitude = NULL
                  renderVersion = 4
                supportedStatus = 0
                   videoInPoint = NULL
                  videoOutPoint = NULL
          videoPosterFramePoint = NULL
                  showInLibrary = 1
                      editState = NULL
                 contentVersion = NULL
              propertiesVersion = NULL
             faceDetectionState = 4
     faceDetectionIsFromPreview = 0
faceDetectionRotationFromMaster = NULL
                    hasKeywords = NULL
                        subType = 32
                    specialType = 0
                     momentUuid = H8M6+2XgSYaziE9rUY2WdQ
                  burstPickType = 0
            extendedDescription = NULL
                 outputUpToDate = 1
         previewsAdjustmentUuid = UNADJUSTEDNONRAW
          pendingAdjustmentUuid = NULL
             faceAdjustmentUuid = UNADJUSTEDNONRAW
                 lastSharedDate = NULL
           videoCpDurationValue = NULL
       videoCpDurationTimescale = NULL
       videoCpImageDisplayValue = NULL
   videoCpImageDisplayTimescale = NULL
         videoCpVisibilityState = NULL
      colorSpaceValidationState = 2
                  momentSortIdx = 0
```

About a third of the way through the 76! fields there's `masterUuid`. A quick test proves this can be used as the join key to get the file path we've been looking for all along. Since this is a longer query we pipe it via `stdin` which sqlite reads if no commands are passed.

```
$ cat q.sql
SELECT v.uuid, v.masterUuid, m.imagePath
FROM RKVersion v
JOIN RKMaster m ON m.uuid = v.masterUuid
WHERE v.uuid = '49RT+c2xRB6yWYs8attNuA'
$ sqlite3 -line Library.apdb < q.sql
      uuid = 49RT+c2xRB6yWYs8attNuA
masterUuid = xoCatkSMRueQxFJSb1kxpw
 imagePath = 2016/11/01/20161101-021106/DSC00004.JPG
```

Now that we understand the schema, we can piece it all together. There's one last problem to solve, joining records across two database files. AFAIK, this isn't possible in sqlite directly. One approach is using the `.dump` command on each database and concatenating the output SQL. That can be piped back into sqlite to create a new merged database file. This could be slow depending on the size of your photo library and machine specs.

```
$ sqlite3 Person.db | grep -v sqlite_master >person.sql
$ sqlite3 Library.apdb | grep -v sqlite_master >library.sql
$ cat person.sql library.sql | sqlite merged.db
```

Armed with the unified database we can get a table of images, sizes and face locations.

```
$ cat faces.sql
SELECT m.imagePath, f.sourceWidth, f.sourceHeight, f.centerX, f.centerY, f.size
FROM RKFace f
JOIN RKVersion v ON v.uuid = f.imageId
JOIN RKMaster m ON m.uuid = v.masterUuid
$ sqlite3 -csv merged.db < faces.sql | head
imagePath,sourceWidth,sourceHeight,centerX,centerY,size
2016/11/01/20161101-021106/DSC00004.JPG,6000,4000,0.397666666666667,0.39475,0.108
2016/11/01/20161101-021106/DSC00004.JPG,6000,4000,0.6905,0.4475,0.222
2016/11/01/20161101-021106/DSC00005.JPG,6000,4000,0.519666666666667,0.603,0.375333333333333
2016/11/01/20161101-021106/DSC00006.JPG,6000,4000,0.1505,0.64425,0.162666666666667
2016/11/01/20161101-021106/DSC00006.JPG,6000,4000,0.341166666666667,0.503,0.149666666666667
2016/11/01/20161101-021106/DSC00007.JPG,4000,6000,0.554,0.6655,0.301
2016/11/01/20161101-021106/DSC00008.JPG,6000,4000,0.379333333333333,0.50375,0.308
2016/11/01/20161101-021106/DSC00009.JPG,4000,6000,0.471,0.517666666666667,0.599666666666667
2016/11/01/20161101-021106/DSC00011.JPG,6000,4000,0.405833333333333,0.4635,0.261
```

You can see that a couple of these images have two faces detected in them. I'm planning a follow-up post that'll go into more detailed usage of this data. It will revolve around the [`phace`][phace] package mentioned in the intro. The goal of that tool is to support automatic extraction of facial recognition data and the option to embed it directly in image files.

Hopefully you've learned something new about "reverse engineering" unknown data formats. It can be roughly summarized in 3 easy steps:

  1. `ls`, `tree`, `find`, etc. around directory
  2. `file` on "interesting" or "unrecognized" file(s)
  3. `sqlite`, `xxv`,  or other custom tool to inspect above file(s)

That last step is where all the interesting bits occur.


[^coords]: This intentionally ignores what corner is the origin. It can vary across graphics contexts and may be further impacted by orientation metadata in images. This will be a topic of the follow up when we get into drawing boxes on images around faces.

[^uuids]: Those of you familiar with [UUIDs][wiki-uuid] may be wondering about the format here. Generally these are serialized in the form of `{123e4567-e89b-12d3-a456-426655440000}`. Another approach is to sometimes base64 encode (w/out padding) the raw 16 bytes for string representation. That produces a 22 character length string as we see here.

[phace]: https://github.com/neilpa/phace
[wiki-uuid]: https://wikipedia.org/wiki/uuid
