package pelican

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"unicode"

	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/css"
	"golang.org/x/net/html"
)

const previewCSP = "default-src 'none'; script-src 'none'; style-src 'unsafe-inline'; connect-src 'none'; img-src 'none'; font-src 'none'; media-src 'none'; object-src 'none'; frame-src 'none'; base-uri 'none'; form-action 'none'"

var errPreview = errors.New("unsupported preview content")

func wordSet(words string) map[string]bool {
	m := map[string]bool{}
	for _, w := range strings.Fields(words) {
		m[w] = true
	}
	return m
}

var allowedElements = wordSet(`html head body title style div span main section article header footer p h1 h2 h3 h4 strong em b i small br hr
 svg g defs symbol use path rect circle ellipse line polyline polygon text tspan textpath desc
 lineargradient radialgradient stop clippath mask pattern filter feblend fecolormatrix fecomponenttransfer fecomposite
 feconvolvematrix fediffuselighting fedisplacementmap fedistantlight fedropshadow feflood fefunca fefuncb fefuncg fefuncr
 fegaussianblur femerge femergenode femorphology feoffset fepointlight fespecularlighting fespotlight fetile feturbulence
 animate animatetransform animatemotion mpath`)

var allowedAttributes = wordSet(`id class style lang dir role aria-label aria-hidden aria-labelledby aria-describedby
 width height viewbox preserveaspectratio x y x1 x2 y1 y2 cx cy r rx ry d points dx dy rotate transform transform-origin
 fill fill-opacity fill-rule stroke stroke-width stroke-linecap stroke-linejoin stroke-miterlimit stroke-dasharray stroke-dashoffset stroke-opacity opacity
 clip-path clip-rule mask filter vector-effect paint-order color display visibility overflow
 font-family font-size font-weight font-style text-anchor dominant-baseline alignment-baseline letter-spacing textlength lengthadjust
 gradientunits gradienttransform spreadmethod offset stop-color stop-opacity fx fy fr patternunits patterncontentunits patterntransform
 clippathunits maskunits maskcontentunits filterunits primitiveunits in in2 result mode type values value operator k1 k2 k3 k4
 stddeviation edgemode flood-color flood-opacity lighting-color scale xchannelselector ychannelselector
 kernelmatrix kernelunitlength order divisor bias targetx targety preservealpha surfacescale diffuseconstant specularconstant specularexponent
 azimuth elevation z pointsatx pointsaty pointsatz limitingconeangle basefrequency numoctaves seed stitchtiles radius slope intercept amplitude exponent tablevalues
 attributename attributetype from to by dur begin end repeatcount repeatdur calcmode keytimes keysplines keypoints additive accumulate path keytimes href`)

var animatedAttributes = wordSet(`transform d x y x1 x2 y1 y2 cx cy r rx ry width height dx dy rotate points opacity fill fill-opacity stroke stroke-width stroke-opacity stroke-dashoffset stroke-dasharray offset stop-color stop-opacity`)

func safeFragment(s string) bool {
	if len(s) < 2 || s[0] != '#' {
		return false
	}
	for _, r := range s[1:] {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '-' && r != '.' {
			return false
		}
	}
	return true
}

// A CSS lexer handles comments and token boundaries. Escaped tokens are rejected
// rather than attempting a second, subtly different interpretation from browsers.
func safeCSS(s string) bool {
	if strings.ContainsAny(s, "\\\x00") {
		return false
	}
	l := css.NewLexer(parse.NewInputString(s))
	for {
		tt, data := l.Next()
		token := strings.ToLower(string(data))
		if tt == css.ErrorToken {
			return l.Err() == io.EOF
		}
		switch tt {
		case css.BadStringToken, css.BadURLToken:
			return false
		case css.AtKeywordToken:
			if token != "@keyframes" && token != "@-webkit-keyframes" && token != "@media" && token != "@supports" {
				return false
			}
		case css.URLToken:
			if !strings.HasPrefix(token, "url(") || !strings.HasSuffix(token, ")") {
				return false
			}
			value := strings.Trim(strings.TrimSpace(string(data[4:len(data)-1])), "\"'")
			if !safeFragment(value) {
				return false
			}
		case css.FunctionToken:
			switch token {
			case "url(", "expression(", "image(", "image-set(", "-webkit-image-set(", "src(", "element(", "-moz-element(":
				return false
			}
		case css.IdentToken:
			if token == "behavior" || token == "-moz-binding" {
				return false
			}
		}
	}
}

func preparePreview(raw string) (string, error) {
	if len(raw) > MaxArtifactBytes {
		return "", errPreview
	}
	s := strings.TrimSpace(raw)
	if strings.HasPrefix(s, "```") {
		line := strings.IndexByte(s, '\n')
		if line < 0 || !strings.HasSuffix(s, "```") {
			return "", errPreview
		}
		header := strings.TrimSpace(s[3:line])
		if header != "" && header != "html" && header != "svg" {
			return "", errPreview
		}
		s = strings.TrimSpace(s[line+1 : len(s)-3])
	}
	if !strings.HasPrefix(s, "<") {
		return "", errPreview
	}
	doc, err := html.Parse(strings.NewReader(s))
	if err != nil {
		return "", errPreview
	}
	var head *html.Node
	svgCount, nodes := 0, 0
	var walk func(*html.Node, int) error
	walk = func(n *html.Node, depth int) error {
		nodes++
		if depth > 128 || nodes > 20000 {
			return errPreview
		}
		if n.Type == html.ElementNode {
			tag := strings.ToLower(n.Data)
			if tag == "meta" {
				// Keep only harmless metadata; application supplies the actual values.
				for _, a := range n.Attr {
					if a.Key != "charset" && !(a.Key == "name" && a.Val == "viewport") && a.Key != "content" {
						return errPreview
					}
				}
				n.Parent.RemoveChild(n)
				return nil
			}
			if !allowedElements[tag] {
				return errPreview
			}
			if tag == "head" {
				head = n
			}
			if tag == "svg" {
				svgCount++
			}
			attrs := make([]html.Attribute, 0, len(n.Attr))
			for _, a := range n.Attr {
				key := strings.ToLower(a.Key)
				if key == "xmlns" || a.Namespace == "xmlns" {
					continue
				}
				if a.Namespace != "" && !(a.Namespace == "xlink" && key == "href") {
					return errPreview
				}
				if !allowedAttributes[key] {
					return errPreview
				}
				if key == "href" && !safeFragment(a.Val) {
					return errPreview
				}
				if key == "attributename" && !animatedAttributes[strings.ToLower(a.Val)] {
					return errPreview
				}
				if key == "begin" || key == "end" {
					// No event/syncbase triggers: only numeric clock offsets and indefinite.
					for _, v := range strings.Split(a.Val, ";") {
						v = strings.TrimSpace(v)
						if v == "indefinite" {
							continue
						}
						for _, r := range v {
							if !strings.ContainsRune("0123456789.+-:msh", r) {
								return errPreview
							}
						}
					}
				}
				if key == "style" || key == "fill" || key == "stroke" || key == "filter" || key == "clip-path" || key == "mask" || key == "values" || key == "from" || key == "to" || key == "by" {
					if !safeCSS(a.Val) {
						return errPreview
					}
				}
				attrs = append(attrs, a)
			}
			n.Attr = attrs
			if tag == "style" {
				var content strings.Builder
				for child := n.FirstChild; child != nil; child = child.NextSibling {
					if child.Type != html.TextNode {
						return errPreview
					}
					content.WriteString(child.Data)
				}
				if !safeCSS(content.String()) {
					return errPreview
				}
			}
		}
		for child := n.FirstChild; child != nil; {
			next := child.NextSibling
			if err := walk(child, depth+1); err != nil {
				return err
			}
			child = next
		}
		return nil
	}
	if err = walk(doc, 0); err != nil || svgCount == 0 || head == nil {
		return "", errPreview
	}
	csp := &html.Node{Type: html.ElementNode, Data: "meta", Attr: []html.Attribute{{Key: "http-equiv", Val: "Content-Security-Policy"}, {Key: "content", Val: previewCSP}}}
	head.InsertBefore(csp, head.FirstChild)
	var out bytes.Buffer
	if err = html.Render(&out, doc); err != nil || out.Len() > MaxArtifactBytes {
		return "", errPreview
	}
	return out.String(), nil
}
