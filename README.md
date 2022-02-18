# PDFToImages
Convert PDF to multiple images.

### Preparation before installation
```bash
sudo apt install libmagic-dev libmagickwand-dev
```
##### Check if pkg-config is able to find the right ImageMagick include and libs:
```bash
pkg-config --cflags --libs MagickWand
```
##### Then: 

```bash
export CGO_CFLAGS_ALLOW='-Xpreprocessor'
go get gopkg.in/gographics/imagick.v2/imagick
```
##### Now you can run the program
------------


### When you run the program, you may encounter the following error：
```bash
ImagickException attempt to perform an operation not allowed by the security policy `PDF' @error/constitute.c/IsCoderAuthorized/408
```

Parsing PDF was disabled in /etc/ImageMagick-x/policy.xml due to its inherent insecurity. The same thing did Ubuntu and perhaps more distros will follow as this is recommendation from security researchers.

You may enable it locally by removing 'PDF' from below line:
<policy domain="coder" rights="none" pattern="{PS,PS2,PS3,EPS,PDF,XPS}" />

##### [stackoverflow.com/questions/63988719](https://stackoverflow.com/questions/63988719/attempt-to-perform-an-operation-not-allowed-by-the-security-policy-pdf-error "stackoverflow.com/questions/63988719") may help you .