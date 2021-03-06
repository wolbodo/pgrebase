package main

import (
	"io/ioutil"
	"fmt"
	"regexp"
)

/*
 * Load or reload all views found in FS.
 */
func LoadViews() ( err error ) {
	successfulCount := len( Cfg.ViewFiles )
	errors := make( []string, 0 )
	bypass := make(map[string]bool)

	files, err := ResolveDependencies( Cfg.ViewFiles, Cfg.SqlDirPath + "views" )
	if err != nil { return err }

	views := make( []*View, 0 )
	for i := len( files ) - 1 ; i >= 0 ; i-- {
		file := files[ i ]
		view := View{}
		view.Path = file
		views = append( views, &view )

		err = DownPass( &view, view.Path )
		if err != nil {
			successfulCount--
			errors = append( errors, fmt.Sprintf( "%v\n", err ) )
			bypass[ view.Path ] = true
		}
	}

	for i := len( views ) - 1 ; i >= 0 ; i-- {
		view := views[ i ]
		if _, ignore := bypass[ view.Path ] ; ! ignore {
			err = UpPass( view, view.Path )
			if err != nil {
				successfulCount--
				errors = append( errors, fmt.Sprintf( "%v\n", err ) )
			}
		}
	}

	Report( "views", successfulCount, len( Cfg.ViewFiles ), errors )

	return
}

type View struct {
	CodeUnit
}

/*
 * Load view definition from file
 */
func ( view *View ) Load() ( err error ) {
	definition, err := ioutil.ReadFile( view.Path )
	if err != nil { return err }
	view.Definition = string( definition )

	return
}

/*
 * Parse view for name
 */
func ( view *View ) Parse() ( err error ) {
	nameFinder := regexp.MustCompile( `(?is)CREATE(?:\s+OR\s+REPLACE)?\s+VIEW\s+(\S+)` )
	subMatches := nameFinder.FindStringSubmatch( view.Definition )

	if len( subMatches ) < 2 {
		return fmt.Errorf( "Can't find a view in %s", view.Path )
	}

	view.Name = subMatches[1]

	return
}

/*
 * Drop existing view from pg
 */
func ( view *View ) Drop() ( err error ) {
	return view.CodeUnit.Drop( `DROP VIEW IF EXISTS ` + view.Name )
}

/*
 * Create the view in pg
 */
func ( view *View ) Create() ( err error ) {
	return view.CodeUnit.Create( view.Definition )
}
